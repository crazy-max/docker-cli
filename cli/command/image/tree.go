package image

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/containerd/platforms"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/streams"
	"github.com/docker/docker/api/types/filters"
	imagetypes "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/pkg/stringid"
	"github.com/docker/go-units"
	"github.com/morikuni/aec"
)

type treeOptions struct {
	all     bool
	filters filters.Args
}

func runTree(ctx context.Context, dockerCLI command.Cli, opts treeOptions) error {
	images, err := dockerCLI.Client().ImageList(ctx, imagetypes.ListOptions{
		All:       opts.all,
		Filters:   opts.filters,
		Manifests: true,
	})
	if err != nil {
		return err
	}

	view := make([]topImage, 0, len(images))
	for _, img := range images {
		details := imageDetails{
			ID:        img.ID,
			DiskUsage: units.HumanSizeWithPrecision(float64(img.Size), 3),
			Used:      img.Containers > 0,
		}

		var totalContent int64
		children := make([]subImage, 0, len(img.Manifests))
		for _, im := range img.Manifests {
			if im.Kind != imagetypes.ManifestKindImage {
				continue
			}

			im := im
			sub := subImage{
				Platform:  platforms.Format(im.ImageData.Platform),
				Available: im.Available,
				Details: imageDetails{
					ID:          im.ID,
					DiskUsage:   units.HumanSizeWithPrecision(float64(im.Size.Total), 3),
					Used:        len(im.ImageData.Containers) > 0,
					ContentSize: units.HumanSizeWithPrecision(float64(im.Size.Content), 3),
				},
			}

			totalContent += im.Size.Content
			children = append(children, sub)
		}

		details.ContentSize = units.HumanSizeWithPrecision(float64(totalContent), 3)
		view = append(view, topImage{
			Names:    img.RepoTags,
			Details:  details,
			Children: children,
		})
	}

	return printImageTree(dockerCLI, view)
}

type imageDetails struct {
	ID          string
	DiskUsage   string
	Used        bool
	ContentSize string
}

type topImage struct {
	Names    []string
	Details  imageDetails
	Children []subImage
}

type subImage struct {
	Platform  string
	Available bool
	Details   imageDetails
}

const columnSpacing = 3

func printImageTree(dockerCLI command.Cli, images []topImage) error {
	out := dockerCLI.Out()
	_, width := out.GetTtySize()
	if width == 0 {
		width = 80
	}
	if width < 20 {
		width = 20
	}

	headerColor := aec.NewBuilder(aec.DefaultF, aec.Bold).ANSI
	topNameColor := aec.NewBuilder(aec.BlueF, aec.Underline, aec.Bold).ANSI
	normalColor := aec.NewBuilder(aec.DefaultF).ANSI
	greenColor := aec.NewBuilder(aec.GreenF).ANSI
	if !out.IsTerminal() {
		headerColor = noColor{}
		topNameColor = noColor{}
		normalColor = noColor{}
		greenColor = noColor{}
	}

	columns := []imgColumn{
		{Title: "Image", Width: 0, Left: true},
		{
			Title: "ID",
			Width: 12,
			DetailsValue: func(d *imageDetails) string {
				return stringid.TruncateID(d.ID)
			},
		},
		{
			Title: "Disk usage",
			Width: 10,
			DetailsValue: func(d *imageDetails) string {
				return d.DiskUsage
			},
		},
		{
			Title: "Content size",
			Width: 12,
			DetailsValue: func(d *imageDetails) string {
				return d.ContentSize
			},
		},
		{
			Title: "Used",
			Width: 4,
			Color: &greenColor,
			DetailsValue: func(d *imageDetails) string {
				if d.Used {
					return "✔"
				}
				return " "
			},
		},
	}

	nameWidth := int(width)
	for idx, h := range columns {
		if h.Width == 0 {
			continue
		}
		d := h.Width
		if idx > 0 {
			d += columnSpacing
		}
		// If the first column gets too short, remove remaining columns
		if nameWidth-d < 12 {
			columns = columns[:idx]
			break
		}
		nameWidth -= d
	}

	// Try to make the first column as narrow as possible
	widest := widestFirstColumnValue(columns, images)
	if nameWidth > widest {
		nameWidth = widest
	}
	columns[0].Width = nameWidth

	// Print columns
	for i, h := range columns {
		if i > 0 {
			_, _ = fmt.Fprint(out, strings.Repeat(" ", columnSpacing))
		}

		_, _ = fmt.Fprint(out, h.Print(headerColor, h.Title))
	}

	_, _ = fmt.Fprintln(out)

	// Print images
	for idx, img := range images {
		if idx != 0 {
			_, _ = fmt.Fprintln(out, "")
		}

		printNames(out, columns, img, topNameColor)
		printDetails(out, columns, normalColor, img.Details)
		printChildren(out, columns, img, normalColor)
	}

	return nil
}

func printDetails(out *streams.Out, headers []imgColumn, defaultColor aec.ANSI, details imageDetails) {
	for _, h := range headers {
		if h.DetailsValue == nil {
			continue
		}

		_, _ = fmt.Fprint(out, strings.Repeat(" ", columnSpacing))
		clr := defaultColor
		if h.Color != nil {
			clr = *h.Color
		}
		val := h.DetailsValue(&details)
		_, _ = fmt.Fprint(out, h.Print(clr, val))
	}
	fmt.Printf("\n")
}

func printChildren(out *streams.Out, headers []imgColumn, img topImage, normalColor aec.ANSI) {
	for idx, sub := range img.Children {
		clr := normalColor
		if !sub.Available {
			clr = normalColor.With(aec.Faint)
		}

		if idx != len(img.Children)-1 {
			_, _ = fmt.Fprint(out, headers[0].Print(clr, "├─ "+sub.Platform))
		} else {
			_, _ = fmt.Fprint(out, headers[0].Print(clr, "└─ "+sub.Platform))
		}

		printDetails(out, headers, clr, sub.Details)
	}
}

func printNames(out *streams.Out, headers []imgColumn, img topImage, color aec.ANSI) {
	for nameIdx, name := range img.Names {
		if nameIdx != 0 {
			_, _ = fmt.Fprintln(out, "")
		}
		_, _ = fmt.Fprint(out, headers[0].Print(color, name))
	}
}

type imgColumn struct {
	Title string
	Width int
	Left  bool

	DetailsValue func(*imageDetails) string
	Color        *aec.ANSI
}

func truncateRunes(s string, length int) string {
	runes := []rune(s)
	if len(runes) > length {
		return string(runes[:length-3]) + "..."
	}
	return s
}

func (h imgColumn) Print(clr aec.ANSI, s string) (out string) {
	if h.Left {
		return h.PrintL(clr, s)
	}
	return h.PrintC(clr, s)
}

func (h imgColumn) PrintC(clr aec.ANSI, s string) (out string) {
	ln := utf8.RuneCountInString(s)

	if ln > h.Width {
		return clr.Apply(truncateRunes(s, h.Width))
	}

	fill := h.Width - ln

	l := fill / 2
	r := fill - l

	return strings.Repeat(" ", l) + clr.Apply(s) + strings.Repeat(" ", r)
}

func (h imgColumn) PrintL(clr aec.ANSI, s string) string {
	ln := utf8.RuneCountInString(s)
	if ln > h.Width {
		return clr.Apply(truncateRunes(s, h.Width))
	}

	return clr.Apply(s) + strings.Repeat(" ", h.Width-ln)
}

type noColor struct{}

func (a noColor) With(ansi ...aec.ANSI) aec.ANSI {
	return aec.NewBuilder(ansi...).ANSI
}

func (a noColor) Apply(s string) string {
	return s
}

func (a noColor) String() string {
	return ""
}

// widestFirstColumnValue calculates the width needed to fully display the image names and platforms.
func widestFirstColumnValue(headers []imgColumn, images []topImage) int {
	width := len(headers[0].Title)
	for _, img := range images {
		for _, name := range img.Names {
			if len(name) > width {
				width = len(name)
			}
		}
		for _, sub := range img.Children {
			pl := len(sub.Platform) + len("└─ ")
			if pl > width {
				width = pl
			}
		}
	}
	return width
}
