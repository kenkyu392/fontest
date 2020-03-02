package fontest

import (
	"bytes"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"golang.org/x/text/unicode/runenames"
)

var nlreg = regexp.MustCompile(`\r\n|\r|\n`)

const csvFileNameTimeFormat = "2006-01-02_150405.csv"

// Run the fontest
func Run(ctx context.Context, argv []string, outStream, errStream io.Writer) error {
	log.SetOutput(errStream)
	log.SetFlags(log.LstdFlags)
	log.SetPrefix("[fontest] ")

	fs := flag.NewFlagSet("fontest", flag.ContinueOnError)
	fs.Usage = func() {
		name := fs.Name()
		out := fs.Output()
		fmt.Fprint(out, "Fontest is a CLI tool for checking the characters included in font files.\n")
		fmt.Fprintf(out, "Usage: %s <options> ... <files> ...\n", name)
		fs.PrintDefaults()
	}
	fs.SetOutput(errStream)

	hf := fs.Bool("help", false, "Prints the help.")
	vf := fs.Bool("version", false, "Prints the version.")
	ff := fs.String("file", "", "Characters file.")

	if err := fs.Parse(argv); err != nil {
		return err
	}

	if *hf {
		fs.Usage()
		return flag.ErrHelp
	}

	if *vf {
		_, err := fmt.Fprintf(outStream, "v%s\n", version)
		return err
	}

	if *ff == "" {
		fs.Usage()
		return flag.ErrHelp
	}

	inputRaw, err := ioutil.ReadFile(*ff)
	if err != nil {
		return fmt.Errorf("%s: No such file or directory", *ff)
	}
	var (
		runes   = bytes.Runes(nlreg.ReplaceAll(inputRaw, []byte{}))
		lines   = make([]string, 0)
		records = make([][]string, len(runes)+1)
	)
	div := 26
	for i := 0; i < len(runes); i += div {
		if i+div < len(runes) {
			lines = append(lines, string(runes[i:(i+div)]))
		} else {
			lines = append(lines, string(runes[i:]))
		}
	}

	records[0] = []string{"Character", "Unicode Name", "Unicode Point"}
	sort.Slice(runes, func(i, j int) bool {
		return runes[i] < runes[j]
	})
	for i, r := range runes {
		records[i+1] = []string{
			fmt.Sprintf("%#U", r),
			runenames.Name(r),
			fmt.Sprintf("%U", r),
		}
	}

	for _, fontFile := range fs.Args() {
		fontName := getFontName(fontFile)
		ttf, err := loadFont(fontFile)
		if err != nil {
			return err
		}
		records[0] = append(records[0], fontName)
		for i, r := range runes {
			records[i+1] = append(records[i+1], fmt.Sprintf("%v", ttf.Index(r) != 0))
		}
		if err := saveImage(ttf, fontName, lines); err != nil {
			return err
		}
	}
	if err := saveResult(records); err != nil {
		return err
	}
	return nil
}

func saveImage(ttf *truetype.Font, title string, lines []string) error {
	var fg, bg = image.White, image.Black
	const dpi = 72              // screen resolution in Dots Per Inch
	const size = 16             // font size in points
	const spacing = 1.5         // line spacing (e.g. 2 means double spaced)
	const imgW, imgH = 595, 842 // A4 paper size
	rgba := image.NewRGBA(image.Rect(0, 0, imgW, imgH))
	draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)

	d := &font.Drawer{
		Dst: rgba,
		Src: fg,
		Face: truetype.NewFace(ttf, &truetype.Options{
			Size:    size,
			DPI:     dpi,
			Hinting: font.HintingFull,
		}),
	}

	y := 10 + int(math.Ceil(size*dpi/72))
	dy := int(math.Ceil(size * spacing * dpi / 72))
	d.Dot = fixed.Point26_6{
		X: (fixed.I(imgW) - d.MeasureString(title)) / 2,
		Y: fixed.I(y),
	}
	d.DrawString(title)
	y += dy
	for _, line := range lines {
		d.Dot = fixed.P(10, y)
		d.DrawString(line)
		y += dy
	}

	buf := new(bytes.Buffer)
	if err := png.Encode(buf, rgba); err != nil {
		return err
	}
	if err := ioutil.WriteFile("__"+title+".png", buf.Bytes(), os.ModePerm); err != nil {
		return err
	}
	return nil
}

func saveResult(records [][]string) error {
	var buf = bytes.NewBuffer([]byte{0xEF, 0xBB, 0xBF})
	if err := csv.NewWriter(buf).WriteAll(records); err != nil {
		return err
	}
	if err := ioutil.WriteFile(time.Now().Format(csvFileNameTimeFormat), buf.Bytes(), os.ModePerm); err != nil {
		return err
	}
	return nil
}

func loadFont(name string) (*truetype.Font, error) {
	raw, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("%s: No such file or directory", name)
	}
	ttf, err := truetype.Parse(raw)
	if err != nil {
		return nil, err
	}
	return ttf, nil
}

func getFontName(path string) string {
	return filepath.Base(path[:len(path)-len(filepath.Ext(path))])
}
