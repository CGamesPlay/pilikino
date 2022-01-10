package notedb

import (
	"net/url"
	"path/filepath"
)

// DetectResult is used to indicate how likely a given URL is to point to a
// database handled by the queried format.
type DetectResult int

const (
	DetectResultNegative = DetectResult(iota - 1)
	DetectResultUnknown
	DetectResultPositive
)

// FormatDescription is used to register a database format with Pilikino.
type FormatDescription struct {
	// ID is the internal ID of the format. It is also used as a URI scheme
	// prefix, so it should be alphanumeric with dashes.
	ID string
	// Description is a short human-readable description of this format.
	Description string
	// Documentation is the human-readable documentation for this format. It
	// should include information about how to locate or produce the database,
	// limitations of the parser, or any other useful information to the user.
	Documentation string
	// Open should load and return the database at the provided URL.
	Open func(dbURL *url.URL) (Database, error)
	// Detect should examine the URL and determine if it is likely to be a
	// database handled by this format. This process should only examine the
	// URL and not the content pointed to by the URL. The format which is "most
	// confident" about the detection is the one which will be selected. The
	// default implementation returns DetectResultNegative.
	Detect func(dbURL *url.URL) DetectResult
}

var registeredFormats map[string]*FormatDescription

func init() {
	registeredFormats = make(map[string]*FormatDescription)
}

// RegisterFormat registers the given format for use.
func RegisterFormat(fmt FormatDescription) {
	registeredFormats[fmt.ID] = &fmt
}

// ResolveURL resolves the provided path into a full database URL.
func ResolveURL(path string) (*url.URL, error) {
	dbURL, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	if dbURL.Scheme == "" {
		if path[0] != '/' {
			path, err = filepath.Abs(path)
			if err != nil {
				return nil, err
			}
		}
		dbURL = &url.URL{Path: path}

		var bestFormat *FormatDescription
		bestConfidence := DetectResultNegative
		for _, fmt := range registeredFormats {
			if fmt.Detect == nil {
				continue
			}
			confidence := fmt.Detect(dbURL)
			if confidence > bestConfidence {
				bestFormat = fmt
				bestConfidence = confidence
			}
		}

		if bestFormat != nil {
			dbURL.Scheme = bestFormat.ID
		} else {
			dbURL.Scheme = "file"
		}
	}

	return dbURL, nil
}
