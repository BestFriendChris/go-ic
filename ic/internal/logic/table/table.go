package table

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

func PrintTable(w io.Writer, table any) error {
	valType := reflect.TypeOf(table)
	if valType.Kind() != reflect.Slice {
		return fmt.Errorf("must be a slice: got %s", valType.Kind())
	}
	elemType := valType.Elem()
	if elemType.Kind() == reflect.Pointer {
		elemType = elemType.Elem()
	}
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("must be a slice of structs: got slice of %s", elemType.Kind())
	}

	slc := reflect.ValueOf(table)

	output := make([]string, 2+(slc.Len()*2))

	allStrings := stringifyTableValues(slc)
	widths := colWidths(allStrings)

	addHeader(&output, allStrings[0], widths)
	addRows(&output, allStrings[1:], widths)

	_, err := w.Write([]byte(strings.Join(output, "\n") + "\n"))
	if err != nil {
		return err
	}
	return nil
}

func addHeader(output *[]string, headers []string, widths []int) {
	if len(*output) < 2 {
		panic("output too small for input; must be at least len(2)")
	}

	if len(headers) != len(widths) {
		panic(fmt.Sprintf("headers (len %d) and widths (len %d) not same length", len(headers), len(widths)))
	}

	o := *output
	o[0] = "   |"
	o[1] = "---+"
	for i, header := range headers {
		width := widths[i]
		o[0] += colWithWidth(header, width)
		o[1] += colSep(width)
	}
}

func addRows(output *[]string, rows [][]string, widths []int) {
	const headerOffset = 2
	o := *output
	if len(o) < len(rows)*2+headerOffset {
		panic(fmt.Sprintf("output too small for input; must be at least len(%d)", len(rows)*2+headerOffset))
	}

	for i, row := range rows {
		outputIdx := (i * 2) + headerOffset
		lineNo := i + 1
		if lineNo < 100 {
			o[outputIdx] = fmt.Sprintf("%2d |", lineNo)
		} else {
			o[outputIdx] = fmt.Sprintf("%03d|", lineNo%1000)
		}
		o[outputIdx+1] = colSep(1)
		if len(row) != len(widths) {
			panic(fmt.Sprintf("row[%d] (len %d) and widths (len %d) not same length", i, len(row), len(widths)))
		}
		for colIdx, cell := range row {
			width := widths[colIdx]
			o[outputIdx] += colWithWidth(cell, width)
			o[outputIdx+1] += colSep(width)
		}
	}
}

func colWithWidth(s string, width int) string {
	if len(s) > width {
		panic(fmt.Sprintf("%q (len %d) is shorter than width %d", s, len(s), width))
	}
	return fmt.Sprintf(" %s%s |", s, strings.Repeat(" ", width-len(s)))
}

func colSep(width int) string {
	return fmt.Sprintf("-%s-+", strings.Repeat("-", width))
}

func stringifyTableValues(slcVal reflect.Value) [][]string {
	valType := slcVal.Type()
	if valType.Kind() != reflect.Slice {
		panic(fmt.Sprintf("must be slice: got %s", valType.Kind()))
	}
	elemType := valType.Elem()
	if elemType.Kind() == reflect.Pointer {
		elemType = elemType.Elem()
	}
	if elemType.Kind() != reflect.Struct {
		panic(fmt.Sprintf("must be slice of structs: got slice of %s", elemType.Kind()))
	}

	output := make([][]string, slcVal.Len()+1)
	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		if field.IsExported() {
			output[0] = append(output[0], field.Name)
		}
	}
	for slcIdx := 0; slcIdx < slcVal.Len(); slcIdx++ {
		structVal := slcVal.Index(slcIdx)
		if structVal.Kind() == reflect.Pointer {
			structVal = structVal.Elem()
		}
		for strIdx := 0; strIdx < structVal.NumField(); strIdx++ {
			field := elemType.Field(strIdx)
			if !field.IsExported() {
				continue
			}
			value := structVal.Field(strIdx).Interface()

			// Stringify if possible
			var nextOutput string
			if stringer, ok := value.(fmt.Stringer); ok {
				nextOutput = stringer.String()
			} else {
				nextOutput = fmt.Sprintf("%#v", value)
			}

			output[slcIdx+1] = append(output[slcIdx+1], nextOutput)
		}
	}
	return output
}

func colWidths(data [][]string) []int {
	if len(data) == 0 {
		return make([]int, 0)
	}
	rowLength := len(data[0])
	widths := make([]int, rowLength)
	for rowIdx := 0; rowIdx < len(data); rowIdx++ {
		row := data[rowIdx]
		if len(row) != rowLength {
			panic(fmt.Sprintf("bad row length; expect len(%d) got len(%d)", rowLength, len(row)))
		}
		for colIdx, s := range row {
			if len(s) > widths[colIdx] {
				widths[colIdx] = len(s)
			}
		}
	}
	return widths
}
