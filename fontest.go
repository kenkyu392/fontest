package fontest

import (
	"bytes"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/golang/freetype/truetype"
	"golang.org/x/text/unicode/runenames"
)

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
	inputf := fs.String("input", "", "Characters file.")

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

	if *inputf == "" {
		fs.Usage()
		return flag.ErrHelp
	}

	inputRaw, err := ioutil.ReadFile(*inputf)
	if err != nil {
		return fmt.Errorf("%s: No such file or directory", *inputf)
	}
	var (
		runes   = bytes.Runes(inputRaw)
		records = make([][]string, len(runes)+1)
	)
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
		font, err := readFont(fontFile)
		if err != nil {
			return err
		}
		records[0] = append(records[0], filepath.Base(fontFile))
		for i, r := range runes {
			records[i+1] = append(records[i+1], fmt.Sprintf("%v", font.Index(r) != 0))
		}
	}
	var buf = bytes.NewBuffer([]byte{0xEF, 0xBB, 0xBF})
	if err := csv.NewWriter(buf).WriteAll(records); err != nil {
		return err
	}
	if err := ioutil.WriteFile(time.Now().Format("2006-01-02_150405.csv"), buf.Bytes(), os.ModePerm); err != nil {
		return err
	}
	return nil
}

func readFont(name string) (*truetype.Font, error) {
	raw, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("%s: No such file or directory", name)
	}
	f, err := truetype.Parse(raw)
	if err != nil {
		return nil, err
	}
	return f, nil
}
