// Copyright 2018 PJ Engineering and Business Solutions Pty. Ltd. All rights reserved.

package imports

import (
	"encoding/csv"
	"io"

	dataframe "github.com/rocketlaunchr/dataframe-go"
)

// CSVLoadOptions is likely to change.
type CSVLoadOptions struct {

	// Comma is the field delimiter.
	// The default value is ',' when CSVLoadOption is not provided.
	// Comma must be a valid rune and must not be \r, \n,
	// or the Unicode replacement character (0xFFFD).
	Comma rune

	// Comment, if not 0, is the comment character. Lines beginning with the
	// Comment character without preceding whitespace are ignored.
	// With leading whitespace the Comment character becomes part of the
	// field, even if TrimLeadingSpace is true.
	// Comment must be a valid rune and must not be \r, \n,
	// or the Unicode replacement character (0xFFFD).
	// It must also not be equal to Comma.
	Comment rune

	// If TrimLeadingSpace is true, leading white space in a field is ignored.
	// This is done even if the field delimiter, Comma, is white space.
	TrimLeadingSpace bool

	// LargeDataSet should be set to true for large datasets.
	// It will set the capacity of the underlying slices of the dataframe by performing a basic parse
	// of the full dataset before processing the data fully.
	// Preallocating memory can provide speed improvements. Benchmarks should be performed for your use-case.
	LargeDataSet bool

	// DictateDataType is used to inform LoadFromCSV what the true underlying data type is for a given field name.
	// The value for a given key must be of the data type of the data. For a string use "". For a int64 use int64(0).
	DictateDataType map[string]interface{}
}

// LoadFromCSV will load data from a csv file.
// WARNING: The API may change in the future.
func LoadFromCSV(r io.ReadSeeker, options ...CSVLoadOptions) (*dataframe.DataFrame, error) {

	seriess := []dataframe.Series{}
	var df *dataframe.DataFrame
	var init *dataframe.SeriesInit

	cr := csv.NewReader(r)
	cr.ReuseRecord = true
	if len(options) > 0 {
		cr.Comma = options[0].Comma
		cr.Comment = options[0].Comment
		cr.TrimLeadingSpace = options[0].TrimLeadingSpace

		// Count how many rows we have in order to preallocate underlying slices
		if options[0].LargeDataSet {
			init = &dataframe.SeriesInit{}
			for {
				_, err := cr.Read()
				if err != nil {
					if err == io.EOF {
						r.Seek(0, io.SeekStart)
						break
					}
					return nil, err
				}
				init.Size++
			}
			if init.Size > 0 {
				init.Size-- // Remove the space allocated for the "heading"
			}
		}
	}

	var row int
	for {
		rec, err := cr.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if row == 0 {
			// Create the series
			for _, name := range rec {
				seriess = append(seriess, dataframe.NewSeriesString(name, init))
			}
			df = dataframe.NewDataFrame(seriess...)
		} else {
			vals := []interface{}{}
			for _, v := range rec {
				vals = append(vals, v)
			}
			if init == nil {
				df.Append(vals...)
			} else {
				df.UpdateRow(row-1, vals...)
			}

		}
		row++
	}

	if row < 1 {
		return nil, dataframe.ErrNoRows
	}

	return dataframe.NewDataFrame(seriess...), nil
}
