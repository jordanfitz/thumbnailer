package main

import (
	"bufio"
	"fmt"
	"image/jpeg"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/jordanfitz/thumbnailer"
	"github.com/spf13/cobra"
	"golang.org/x/image/draw"
)

func confirm(actionMessage string) bool {
	fmt.Printf("%s [y/N]: ", actionMessage)

	stdin := bufio.NewReader(os.Stdin)
	response, err := stdin.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	if response = strings.ToLower(strings.TrimSpace(response)); response == "" {
		return false
	}

	return unicode.ToLower(rune(response[0])) == 'y'
}

var Scalers = map[string]draw.Scaler{
	"NearestNeighbor": draw.NearestNeighbor,
	"ApproxBiLinear":  draw.ApproxBiLinear,
	"BiLinear":        draw.BiLinear,
	"CatmullRom":      draw.CatmullRom,
}

var OutFormats = map[string]thumbnailer.OutputFormat{
	"original": thumbnailer.OriginalFormat,
	"jpeg":     thumbnailer.JPG,
	"jpg":      thumbnailer.JPG,
	"png":      thumbnailer.PNG,
}

type Config struct {
	InputFiles   []string
	OutputDir    string
	OutputPrefix string
	OutFormat    string
	MaxSize      int
	Quality      int
	Scaler       string
	Force        bool
}

func (c Config) Validate() error {
	if _, ok := OutFormats[c.OutFormat]; !ok {
		return fmt.Errorf("invalid output format '%s'", c.OutFormat)
	}
	if c.OutputDir == "" && c.OutputPrefix == "" {
		return fmt.Errorf("at least one of output path and output prefix must be set")
	}
	if c.MaxSize < 1 {
		return fmt.Errorf("max-size must be at least 1")
	}
	if c.Quality < 0 || c.Quality > 100 {
		return fmt.Errorf("jpg quality must be between 0 and 100")
	}
	if _, ok := Scalers[c.Scaler]; !ok {
		return fmt.Errorf("invalid scaler '%s'", c.Scaler)
	}
	return nil
}

func execute(c Config) error {
	scaler := Scalers[c.Scaler]
	outFormat := OutFormats[c.OutFormat]

	t := thumbnailer.New().
		With(thumbnailer.OutFormat(outFormat)).
		With(thumbnailer.MaxSize(c.MaxSize)).
		With(thumbnailer.Quality(c.Quality)).
		With(thumbnailer.Scaler(scaler))
	_ = t

	for _, file := range c.InputFiles {
		abs, err := filepath.Abs(file)
		if err != nil {
			return err
		}

		fi, err := os.Stat(abs)
		if err != nil {
			return err
		}
		inputMode := fi.Mode()

		data, err := os.ReadFile(abs)
		if err != nil {
			return err
		}

		outputDir := c.OutputDir
		if outputDir == "" {
			outputDir = path.Dir(abs)
		} else if outputDir, err = filepath.Abs(c.OutputDir); err != nil {
			return err
		}

		outputName := fmt.Sprintf("%s%s", c.OutputPrefix, path.Base(abs))
		if outFormat != thumbnailer.OriginalFormat {
			outputName = strings.TrimSuffix(outputName, path.Ext(outputName))
			outputName += "." + c.OutFormat
		}

		outputPath := path.Join(outputDir, outputName)

		if !c.Force {
			if _, err = os.Stat(outputPath); err != nil && !os.IsNotExist(err) {
				return err
			}
			if err == nil && !confirm(
				fmt.Sprintf("%s already exists in the output directory - overwrite?", outputName),
			) {
				continue
			}
		}

		outputData, err := t.With(thumbnailer.Image(data)).Create()
		if err != nil {
			return err
		}
		if err := os.WriteFile(outputPath, outputData, inputMode); err != nil {
			return err
		}

		fmt.Println(abs)
		fmt.Println("  ->", outputPath)
	}

	return nil
}

func main() {
	var c Config

	rootCmd := &cobra.Command{
		Use:   "thumbnailer <image>...",
		Short: "Generate thumbnails for images",
		Args:  cobra.MinimumNArgs(1),
		PreRunE: func(_ *cobra.Command, args []string) error {
			c.InputFiles = args

			if c.OutputDir != "" {
				fs, err := os.Stat(c.OutputDir)
				if err != nil {
					if os.IsNotExist(err) {
						return os.MkdirAll(c.OutputDir, 0744)
					}
					return err
				}
				if !fs.IsDir() {
					return fmt.Errorf("output path '%s' is not a directory", c.OutputDir)
				}
			}

			return c.Validate()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return execute(c)
		},
	}

	rootCmd.Flags().BoolVar(&c.Force, "force", false, "force overwrite existing files")

	rootCmd.Flags().StringVarP(&c.OutputDir, "output", "o", "",
		"output directory (default same as input file(s))")
	rootCmd.Flags().StringVarP(&c.OutFormat, "format", "f", "original",
		"output format (original/jp[e]g/png)")
	rootCmd.Flags().StringVarP(&c.OutputPrefix, "prefix", "p", "t_",
		"prefix for output file name")
	rootCmd.Flags().IntVarP(&c.MaxSize, "max-size", "m", 300,
		"maximum size for thumbnail images")
	rootCmd.Flags().IntVarP(&c.Quality, "jpg-quality", "j", jpeg.DefaultQuality,
		"quality for JPG output (0-100)")
	rootCmd.Flags().StringVarP(&c.Scaler, "scaler", "s", "ApproxBiLinear",
		"scaler to use when downsizing images (NearestNeighbor/ApproxBiLinear/BiLinear/CatmullRom)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}
