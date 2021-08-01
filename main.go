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

type validationError struct {
	File string
	Err  error
}

var _ error = &validationError{}

func (e *validationError) Error() string {
	return fmt.Sprintf("%s: %s", e.File, e.Err.Error())
}

func (e *validationError) annotateGitHubFile() {
	const errAnnotationFmt = "::error file=%s::%s\n"
	v := strings.ReplaceAll(e.Err.Error(), "%", "%25")
	v = strings.ReplaceAll(v, "\r", "%0D")
	v = strings.ReplaceAll(v, "\n", "%0A")

	fmt.Printf(errAnnotationFmt, e.File, v)
}

var (
	fastlanePath        string
	enableGAAnnotations bool
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
		const errFmt = "failed to read directory %q: %s\n"
		fmt.Fprintf(os.Stderr, errFmt, metadataPath, err)
		os.Exit(1)
	}

	errs := make([]error, 0)
	for _, f := range files {
		if !f.IsDir() {
			// we are only interested in directories
			continue
		}

		localePath := filepath.Join(metadataPath, f.Name())
		imagesPath := filepath.Join(localePath, "images")
		changelogsPath := filepath.Join(localePath, "changelogs")
		errs = append(errs, checkDescriptiveTexts(localePath)...)
		errs = append(errs, checkImages(imagesPath)...)
		errs = append(errs, checkChangelogs(changelogsPath)...)
	}

	fmt.Println("found", len(errs), "errors!")
	for _, err := range errs {
		if ve, ok := err.(*validationError); ok && enableGAAnnotations {
			ve.annotateGitHubFile()
		}

		fmt.Fprintln(os.Stderr, err.Error())
	}

	if len(errs) > 0 {
		os.Exit(1)
	}
}

// checkDescriptiveTexts checks *.txt files in metadata. It returns a slice of
// `error` with all IO and validation errors.
func checkDescriptiveTexts(localePath string) []error {
	descriptiveFileLengths := map[string]int{
		"title.txt":             50,
		"short_description.txt": 80,
		"full_description.txt":  4000,
	}

	errs := make([]error, 0)
	for file, length := range descriptiveFileLengths {
		file = filepath.Join(localePath, file)
		count, err := getCharacterCount(file)
		if err != nil {
			const errFmt = "failed to read file %q: %w"
			errs = append(errs, fmt.Errorf(errFmt, file, err))
		} else if count > length {
			const errFmt = "content length exceeded: expected=%d, got=%d"
			errs = append(errs, &validationError{
				File: file,
				Err:  fmt.Errorf(errFmt, length, count),
			})
		}
	}

	return errs
}

// getCharacterCount counts the utf-8 characters in the given file.
func getCharacterCount(filePath string) (int, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return 0, err
	}

	return utf8.RuneCountInString(strings.TrimSpace(string(content))), nil
}

// checkImages checks image assets in `images/*` including screenshots. It
// returns a slice of `error` with all IO and validation errors.
func checkImages(imagesPath string) []error {
	files, err := ioutil.ReadDir(imagesPath)
	// since directory is optional, ignore 'not exist' errors.
	if err != nil && !os.IsNotExist(err) {
		const errFmt = "failed to read directory %q: %w"
		return []error{fmt.Errorf(errFmt, imagesPath, err)}
	}

	errs := make([]error, 0)
	for _, file := range files {
		if file.IsDir() {
			if strings.HasSuffix(file.Name(), "Screenshots") {
				errs = append(errs, checkScreenshots(filepath.Join(imagesPath, file.Name()))...)
			}

			continue
		}

		filePath := filepath.Join(imagesPath, file.Name())
		config, err := getImageConfig(filePath)
		if err != nil {
			const errFmt = "failed to read image %q: %w"
			errs = append(errs, fmt.Errorf(errFmt, filePath, err))
			continue
		}

		switch name := strings.TrimSuffix(filepath.Base(file.Name()), filepath.Ext(file.Name())); name {
		case "icon":
			if config.width != config.height || config.width != 512 {
				const errFmt = "icon must be 512x512: got=%dx%d"
				errs = append(errs, &validationError{
					File: filePath,
					Err:  fmt.Errorf(errFmt, config.width, config.height),
				})
			}
			if config.format != "png" {
				errs = append(errs, &validationError{
					File: filePath,
					Err:  fmt.Errorf("icon must be a PNG"),
				})
			}
		case "featureGraphic":
			if config.width != 1024 || config.height != 500 {
				const errFmt = "featureGraphic must be 1024x500: got=%dx%d"
				errs = append(errs, &validationError{
					File: filePath,
					Err:  fmt.Errorf(errFmt, config.width, config.height),
				})
			}
			if !config.opaque {
				errs = append(errs, &validationError{
					File: filePath,
					Err:  fmt.Errorf("featureGraphic must be opaque"),
				})
			}
		case "promoGraphic":
			if config.width != 180 || config.height != 120 {
				const errFmt = "promoGraphic must be 180x120: got=%dx%d"
				errs = append(errs, &validationError{
					File: filePath,
					Err:  fmt.Errorf(errFmt, config.width, config.height),
				})
			}
			if !config.opaque {
				errs = append(errs, &validationError{
					File: filePath,
					Err:  fmt.Errorf("promoGraphic must be opaque"),
				})
			}
		case "tvBanner":
			if config.width != 1280 || config.height != 720 {
				const errFmt = "tvBanner must be 1280x720: got=%dx%d"
				errs = append(errs, &validationError{
					File: filePath,
					Err:  fmt.Errorf(errFmt, config.width, config.height),
				})
			}
			if !config.opaque {
				errs = append(errs, &validationError{
					File: filePath,
					Err:  fmt.Errorf("tvBanner must be opaque"),
				})
			}
		}
	}

	return errs
}

// checkScreenshots checks all screenshot images. It returns a slice of `error`
// with all IO and validation errors.
func checkScreenshots(screenshotsPath string) []error {
	files, err := ioutil.ReadDir(screenshotsPath)
	if err != nil {
		const errFmt = "failed to read directory %q: %w"
		return []error{fmt.Errorf(errFmt, screenshotsPath, err)}
	}

	errs := make([]error, 0)
	for _, file := range files {
		imagePath := filepath.Join(screenshotsPath, file.Name())
		config, err := getImageConfig(imagePath)
		if err != nil {
			const errFmt = "failed to read image %q: %w"
			errs = append(errs, fmt.Errorf(errFmt, imagePath, err))
			continue
		}

		if config.width < 320 || config.width > 3840 {
			const errFmt = "width should be in range 320px-3840px: got=%dpx"
			errs = append(errs, &validationError{
				File: imagePath,
				Err:  fmt.Errorf(errFmt, config.width),
			})
		}

		if config.height < 320 || config.height > 3840 {
			const errFmt = "height should be in range 320px-3840px: got=%dpx"
			errs = append(errs, &validationError{
				File: imagePath,
				Err:  fmt.Errorf(errFmt, config.height),
			})
		}

		width := float64(config.width)
		height := float64(config.height)
		ratio := math.Max(width, height) / math.Min(height, width)
		if ratio > 2.0 {
			const errFmt = "'max:min' edge radio should be at most 2.0: got=%.2f"
			errs = append(errs, &validationError{
				File: imagePath,
				Err:  fmt.Errorf(errFmt, ratio),
			})
		}
	}

	return errs
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
			return nil, fmt.Errorf("failed to determine if image is opaque")
		}
	}

	return &imageConfig{
		width:  config.Width,
		height: config.Height,
		opaque: opaque,
		format: format,
	}, nil
}

// checkChangelogs checks `changelogs/*.txt` files in metadata. It returns a
// slice of `error` containing both IO and validation errors.
func checkChangelogs(changelogsPath string) []error {
	files, err := ioutil.ReadDir(changelogsPath)
	// since directory is optional, ignore 'not exist' errors.
	if err != nil && !os.IsNotExist(err) {
		const errFmt = "failed to read directory %q: %w"
		return []error{fmt.Errorf(errFmt, changelogsPath, err)}
	}

	errs := make([]error, 0)
	for _, file := range files {
		if file.IsDir() {
			continue // not expecting one.. but okay...?
		}

		filePath := filepath.Join(changelogsPath, file.Name())
		count, err := getCharacterCount(filePath)
		if err != nil {
			const errFmt = "failed to read file %q: %w"
			errs = append(errs, fmt.Errorf(errFmt, filePath, err))
		}

		const maxContentLength = 500
		if count > maxContentLength {
			const errFmt = "content length exceeded: expected=%d, got=%d"
			errs = append(errs, &validationError{
				File: filePath,
				Err:  fmt.Errorf(errFmt, maxContentLength, count),
			})
		}
	}

	return errs
}
