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
	fastlanePath        string
	enableGAAnnotations bool

	// nErrors keeps the count of failed checks during a single run.
	nErrors uint
)

func init() {
	flag.StringVar(&fastlanePath, "fastlane-path", "./fastlane", "path to the Fastlane directory")
	flag.BoolVar(&enableGAAnnotations, "enable-ga-annotations", false, "enables file annotations for GitHub action")
	flag.Parse()
}

func main() {
	metadataPath := filepath.Join(fastlanePath, "metadata", "android")
	files, err := ioutil.ReadDir(metadataPath)
	if err != nil {
		fmt.Printf("Unable to list metadata directory: %s\n", err.Error())
		os.Exit(1)
	}

	for _, f := range files {
		if !f.IsDir() {
			// we are only interested in directories
			continue
		}

		processDescriptiveTexts(metadataPath, f.Name())
		processImages(metadataPath, f.Name())
		processChangelogs(metadataPath, f.Name())
	}

	if nErrors > 0 {
		fmt.Fprintf(os.Stderr, "%d total checks failing!\n", nErrors)
		os.Exit(1)
	}
}

// processDescriptiveTexts will check *.txt files in metadata. updates the count
// of nErrors for failed checks. also prints all the errors for failing checks
// and annonates corresponding files.
func processDescriptiveTexts(basePath, locale string) {
	const errLengthExceededFmt = "content length exceeded the limit! expected: %d, got: %d"
	descriptiveFileLengths := map[string]int{
		"title.txt":             50,
		"short_description.txt": 80,
		"full_description.txt":  4000,
	}

	for file, length := range descriptiveFileLengths {
		file = filepath.Join(basePath, locale, file)
		count, err := getCharacterCount(file)
		if err == nil && count > length {
			err = fmt.Errorf(errLengthExceededFmt, length, count)
		}

		if err != nil {
			nErrors++
			annotateFileWithError(file, err)
			logError(locale, filepath.Base(file), err)
		}
	}
}

// getCharacterCount counts the utf-8 characters in the given file.
func getCharacterCount(filePath string) (int, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("error reading the file: %s", err.Error())
	}

	return utf8.RuneCountInString(strings.TrimSpace(string(content))), nil
}

// processImages process image assets in images/* including screenshots. updates
// the count of nErrors for failed checks. also prints all the errors for
// failing checks and annonates corresponding files.
func processImages(basePath, locale string) {
	imagesBasePath := filepath.Join(basePath, locale, "images")
	files, err := ioutil.ReadDir(imagesBasePath)
	if err != nil && !os.IsNotExist(err) { // ignore not found errors since this is an optional dir
		nErrors++
		logError(locale, "images", err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			if strings.HasSuffix(file.Name(), "Screenshots") {
				processScreenshotImages(basePath, locale, file.Name())
			}
		} else {
			filePath := filepath.Join(imagesBasePath, file.Name())
			fileName := strings.TrimSuffix(filepath.Base(file.Name()), filepath.Ext(file.Name()))
			config, err := getImageConfig(filePath)
			if err != nil {
				nErrors++
				err = fmt.Errorf("error while reading image: %s", err)
				annotateFileWithError(filePath, err)
				logError(locale, fileName, err)
				continue
			}

			errs := make([]error, 0)
			switch fileName {
			case "icon":
				if config.width != config.height || config.width != 512 {
					errs = append(errs, fmt.Errorf("must be 512x512"))
				}
				if config.format != "png" {
					errs = append(errs, fmt.Errorf("must be a PNG"))
				}
				break
			case "featureGraphic":
				if config.width != 1024 || config.height != 500 {
					errs = append(errs, fmt.Errorf("must be 1024x500"))
				}
				if !config.opaque {
					errs = append(errs, fmt.Errorf("must be opaque"))
				}
				break
			case "promoGraphic":
				if config.width != 180 || config.height != 120 {
					errs = append(errs, fmt.Errorf("must be 180x120"))
				}
				if !config.opaque {
					errs = append(errs, fmt.Errorf("must be opaque"))
				}
				break
			case "tvBanner":
				if config.width != 1280 || config.height != 720 {
					errs = append(errs, fmt.Errorf("must be 1280x720"))
				}
				if !config.opaque {
					errs = append(errs, fmt.Errorf("must be opaque"))
				}
				break
			}

			for _, err = range errs {
				nErrors++
				annotateFileWithError(filePath, err)
				logError(locale, fileName, err)
			}
		}
	}
}

// processScreenshotImages checks all the screenshot images. updates the count
// of nErrors for failed checks. also prints all the errors for failing checks
// and annonates corresponding files.
func processScreenshotImages(basePath, locale, dirName string) {
	path := filepath.Join(basePath, locale, "images", dirName)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		nErrors++
		logError(locale, dirName, err)
		return
	}

	for _, file := range files {
		filePath := filepath.Join(path, file.Name())
		config, err := getImageConfig(filePath)
		if err != nil {
			nErrors++
			annotateFileWithError(filePath, err)
			logError(locale, filepath.Join(dirName, file.Name()), err)
			continue
		}

		if config.width < 320 || config.width > 3840 {
			nErrors++
			err = fmt.Errorf("width should be at least 320px and at most 3840px")
			annotateFileWithError(filePath, err)
			logError(locale, filepath.Join(dirName, file.Name()), err)
		}

		if config.height < 320 || config.height > 3840 {
			nErrors++
			err = fmt.Errorf("height should be at least 320px and at most 3840px")
			annotateFileWithError(filePath, err)
			logError(locale, filepath.Join(dirName, file.Name()), err)
		}

		width := float64(config.width)
		height := float64(config.height)
		if math.Max(width, height)/math.Min(height, width) > 2.0 {
			nErrors++
			err = fmt.Errorf("aspect ratio of max edge to min edge should be at most 2")
			annotateFileWithError(filePath, err)
			logError(locale, filepath.Join(dirName, file.Name()), err)
		}
	}
}

// getImageConfig returns imageConfig for the given image file. returns an error
// it is not able to read the image config.
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

// processDescriptiveTexts will check changelogs/*.txt files in metadata.
// updates the count of nErrors for failed checks. also prints all the errors
// for failing checks and annonates corresponding files.
func processChangelogs(basePath, locale string) {
	const maxContentLength = 500
	const changelogsDirName = "changelogs"
	const errLengthExceededFmt = "content length exceeded the limit! expected: %d, got: %d"

	changelogPath := filepath.Join(basePath, locale, changelogsDirName)
	files, err := ioutil.ReadDir(changelogPath)
	if err != nil && !os.IsNotExist(err) {
		// ignore not found errors since this is an optional dir
		nErrors++
		logError(locale, changelogsDirName, err)
		return
	}
	for _, file := range files {
		if file.IsDir() {
			continue // not expecting one.. but okay...?
		}

		filePath := filepath.Join(changelogPath, file.Name())
		count, err := getCharacterCount(filePath)
		if err == nil && count > maxContentLength {
			err = fmt.Errorf(errLengthExceededFmt, maxContentLength, count)
		}

		if err != nil {
			nErrors++
			annotateFileWithError(filePath, err)
			logError(locale, filepath.Join(changelogsDirName, file.Name()), err)
		}
	}
}

// logError logs the error
func logError(locale, file string, err error) {
	fmt.Fprintf(os.Stderr, "%s/%s: %s\n", locale, file, err.Error())
}

func annotateFileWithError(filePath string, err error) {
	if !enableGAAnnotations {
		return
	}

	const errAnnotationFmt = "::error file=%s::%s\n"
	v := strings.ReplaceAll(err.Error(), "%", "%25")
	v = strings.ReplaceAll(v, "\r", "%0D")
	v = strings.ReplaceAll(v, "\n", "%0A")

	fmt.Printf(errAnnotationFmt, filePath, v)
}
