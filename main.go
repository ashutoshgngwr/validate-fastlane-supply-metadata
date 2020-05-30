package main

import (
	"flag"
	"fmt"
	"image"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	_ "image/jpeg"
	_ "image/png"
)

type imageConfig struct {
	width  int
	height int
	opaque bool
	format string
}

var (
	fastlanePath string
)

func init() {
	flag.StringVar(&fastlanePath, "fastlane-path", "./fastlane", "path to the Fastlane directory")
	flag.Parse()
}

func main() {
	metadataPath := filepath.Join(fastlanePath, "metadata", "android")
	files, err := ioutil.ReadDir(metadataPath)
	if err != nil {
		fmt.Printf("Unable to list metadata directory: %s\n", err.Error())
		os.Exit(1)
	}

	var isErrored error
	for _, f := range files {
		if !f.IsDir() {
			// we are only interested in directories
			continue
		}

		if err := processLocale(metadataPath, f.Name()); err != nil {
			isErrored = err
		}
	}

	if isErrored != nil {
		os.Exit(1)
	}
}

// processLocale will check all metadata items one by one for each locale.
// it returns the last error if the metadata is voilating set guidelines or
// if there was an error while processing any of the files.
// printing of errors is handled internally by each function down the stream.
func processLocale(basePath, locale string) (isErrored error) {
	if err := processDescriptiveTexts(basePath, locale); err != nil {
		isErrored = err
	}

	if err := processImages(basePath, locale); err != nil {
		isErrored = err
	}

	return
}

// processDescriptiveTexts will check *.txt files in metadata. returns an error
// if any of the checks fail. also prints all the errors for failing checks.
func processDescriptiveTexts(basePath, locale string) (isErrored error) {
	descriptiveFileLengths := map[string]int{
		"title.txt":             50,
		"short_description.txt": 80,
		"full_description.txt":  4000,
	}

	for file, length := range descriptiveFileLengths {
		count, err := getCharacterCount(filepath.Join(basePath, locale, file))
		if err == nil && count > length {
			err = fmt.Errorf("content length exceeded the limit got: %d desired: %d", count, length)
		}

		if err != nil {
			isErrored = err
			logError(locale, file, err)
		}
	}

	return
}

// getCharacterCount counts the utf-8 characters in the given file.
func getCharacterCount(filePath string) (int, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("error reading the file: %s", err.Error())
	}

	return utf8.RuneCountInString(strings.TrimSpace(string(content))), nil
}

// processImages process image assets in images/* including screenshots.
// returns an error if any of the checks fail. also prints all the errors
// for failing checks.
func processImages(basePath, locale string) (isErrored error) {
	// ignoring list error because images are optional in metadata
	imagesBasePath := filepath.Join(basePath, locale, "images")
	files, err := ioutil.ReadDir(imagesBasePath)
	if err != nil {
		logError(locale, "images", err)
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			if strings.HasSuffix(file.Name(), "Screenshots") {
				if err := processScreenshotImages(basePath, locale, file.Name()); err != nil {
					isErrored = err
				}
			}
		} else {
			baseName := strings.TrimSuffix(filepath.Base(file.Name()), filepath.Ext(file.Name()))
			config, err := getImageConfig(filepath.Join(imagesBasePath, file.Name()))
			if err != nil {
				logError(locale, baseName, fmt.Errorf("error while reading image: %s", err.Error()))
				isErrored = err
				continue
			}

			switch baseName {
			case "icon":
				if config.width != config.height || config.width != 512 {
					err = fmt.Errorf("must be 512x512")
					logError(locale, baseName, err)
				}
				if config.format != "png" {
					err = fmt.Errorf("must be a PNG")
					logError(locale, baseName, err)
				}
				break
			case "featureGraphic":
				if config.width != 1024 || config.height != 500 {
					err = fmt.Errorf("must be 1024x500")
					logError(locale, baseName, err)
				}
				if !config.opaque {
					err = fmt.Errorf("must be opaque")
					logError(locale, baseName, err)
				}
				break
			case "promoGraphic":
				if config.width != 180 || config.height != 120 {
					err = fmt.Errorf("must be 180x120")
					logError(locale, baseName, err)
				}
				if !config.opaque {
					err = fmt.Errorf("must be opaque")
					logError(locale, baseName, err)
				}
				break
			case "tvBanner":
				if config.width != 1280 || config.height != 720 {
					err = fmt.Errorf("must be 1280x720")
					logError(locale, baseName, err)
				}
				if !config.opaque {
					err = fmt.Errorf("must be opaque")
					logError(locale, baseName, err)
				}
				break

			}

			if err != nil {
				isErrored = err
			}
		}
	}

	return
}

// processScreenshotImages checks all the screenshot images. returns an error if
// any of the checks fail. also prints the errors for all the failing checks.
func processScreenshotImages(basePath, locale, dirName string) (isErrored error) {
	path := filepath.Join(basePath, locale, "images", dirName)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		logError(locale, dirName, err)
		return err
	}

	for _, file := range files {
		config, err := getImageConfig(filepath.Join(path, file.Name()))
		if err != nil {
			logError(locale, filepath.Join(dirName, file.Name()), err)
			isErrored = err
			continue
		}

		if config.width < 320 || config.width > 3840 {
			err = fmt.Errorf("width should be at least 320px and at most 3840px")
			logError(locale, filepath.Join(dirName, file.Name()), err)
		}

		if config.height < 320 || config.height > 3840 {
			err = fmt.Errorf("height should be at least 320px and at most 3840px")
			logError(locale, filepath.Join(dirName, file.Name()), err)
		}

		width := float64(config.width)
		height := float64(config.height)
		if math.Max(width, height)/math.Min(height, width) > 2.0 {
			err = fmt.Errorf("aspect ratio of max edge to min edge should be at most 2")
			logError(locale, filepath.Join(dirName, file.Name()), err)
		}

		if err != nil {
			isErrored = err
		}
	}

	return
}

// getImageConfig returns imageConfig for the given image file. returns
// an error it is not able to read the image config.
func getImageConfig(filePath string) (*imageConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	config, format, err := image.DecodeConfig(file)
	if err != nil {
		return nil, err
	}

	opaque := false
	if format == "png" { // need to check if image is opaque
		if _, err = file.Seek(0, 0); err != nil {
			return nil, err
		}

		image, _, err := image.Decode(file)
		if err != nil {
			return nil, err
		}

		if oimage, ok := image.(interface{ Opaque() bool }); ok {
			opaque = oimage.Opaque()
		} else {
			return nil, fmt.Errorf("unable to determine if image is opaque")
		}
	}

	return &imageConfig{
		width:  config.Width,
		height: config.Height,
		opaque: opaque,
		format: format,
	}, nil
}

// logError logs the error
func logError(locale, file string, err error) {
	fmt.Printf("%s/%s: %s\n", locale, file, err.Error())
}
