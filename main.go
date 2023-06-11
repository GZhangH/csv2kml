package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/fatih/color"
	"golang.org/x/term"
)

var (
	Term_Width int = 0
	Errors     error
)

type DateFrame struct {
	keys []string
	data map[string][]float64
	raw  [][]string
	row  int
	col  int
}

func Int(s string) int {
	v, e := strconv.Atoi(s)
	if e != nil {
		log.Panicln(e)
	}
	return v
}

func Float64(s string) float64 {
	v, e := strconv.ParseFloat(s, 64)
	if e != nil {
		log.Panicln(e)
	}
	return v
}

func Contain[T comparable](s []T, p T) bool {
	for i := range s {
		if s[i] == p {
			return true
		}
	}
	return false
}

func ReadCSV(file string) DateFrame {
	var f *os.File
	var e error
	var r [][]string
	var data = make(map[string][]float64)
	var df DateFrame
	f, e = os.Open(file)
	if e != nil {
		log.Panicln(e)
	}
	defer f.Close()
	r, e = csv.NewReader(f).ReadAll()
	if e != nil {
		log.Panicln(e)
	}
	df.raw = r
	for idy := range r {
		for idx := range r[idy] {
			r[idy][idx] = strings.TrimSpace(r[idy][idx])
		}
	}
	df.keys = r[0]

	for idx := range r[0] {
		data[r[0][idx]] = []float64{}
	}

	for idy := 1; idy < len(r); idy++ {
		for idx := range r[idy] {
			data[r[0][idx]] = append(data[r[0][idx]], Float64(r[idy][idx]))
		}
	}

	df.data = data
	df.row = len(r[0])
	df.col = len(r[1:])

	return df
}

func (df DateFrame) Keys() []string {
	return df.keys
}

func (df DateFrame) Raw() [][]string {
	return df.raw
}

func (df DateFrame) Data() map[string][]float64 {
	return df.data
}

func (df DateFrame) Row() int {
	return df.row
}

func (df DateFrame) Col() int {
	return df.col
}

func KeysMaxLength(df DateFrame) int {
	ms := 0
	for i := range df.keys {
		if len(df.keys[i]) > ms {
			ms = len(df.keys[i])
		}
	}
	return ms
}

func ShowKeys(df DateFrame) {
	line := ""
	ms := KeysMaxLength(df) + 4
	rs := df.Row()/10 + 1
	for i := range df.keys {
		item := fmt.Sprintf("%*d = %#*v, ", rs, i+1, ms, df.keys[i])
		if len(line)+len(item) < Term_Width {
			line += item
		} else {
			fmt.Printf("%s\n", line)
			line = item
		}
		if (i + 1) == df.row {
			fmt.Printf("%s\n", line)
		}
	}
}

const tpl = `<?xml version="1.0" encoding="utf-8" ?>
<kml xmlns="http://www.opengis.net/kml/2.2">
<Document id="root_doc">
<Folder>
    <name>
	{{.FILE}}
    </name>
    <Placemark>
    <name>trajectory</name>
    <description>record trajectory path</description>
    <Style><LineStyle>
        <color>ff0000ff</color></LineStyle><PolyStyle><fill>0</fill>
    </PolyStyle></Style>
        <LineString>
        <coordinates>
		{{.LLA}}
        </coordinates>
        </LineString>
    </Placemark>
    <Placemark>
    <name>start</name>
    <description>start record</description>
        <Point>
        <coordinates>
		{{.LLA_head}}
        </coordinates>
        </Point>
    </Placemark>
    <Placemark>
    <name>end</name>
    <description>end record</description>
        <Point>
        <coordinates>
		{{.LLA_end}}
        </coordinates>
        </Point>
    </Placemark>
</Folder>
</Document></kml>`

func main() {
	Term_Width, _, Errors = term.GetSize(int(os.Stdout.Fd()))
	if Errors != nil {
		log.Panicln(Errors)
	}
	args := os.Args[1:]
	csvfile := ""

	if len(args) == 0 ||
		args[0] == "-h" ||
		args[0] == "--help" {
		fmt.Printf("usage: ./csvkml [csv file]\n")
		return
	} else {
		csvfile = args[0]
	}
	log.Printf("Start\n")

	df := ReadCSV(csvfile)
	fmt.Printf("\nShow keys: \n")
	ShowKeys(df)

	select_lon_idx := 0
	select_lat_idx := 0
	select_alt_idx := 0
	output_kml := ""

	fmt.Printf("\n")
	fmt.Printf("Select the key index of longitude/latitude/altitude\n")

	fmt.Printf("Select the Longitude index: ")
	fmt.Scanf("%d\n", &select_lon_idx)
	for select_lon_idx < 1 || select_lon_idx > df.row {
		fmt.Printf("%s\n",
			color.RedString(
				"Index is over the range, "+
					"please to reselect again..."))
		fmt.Printf("Select the Longitude index: ")
		fmt.Scanf("%d\n", &select_lon_idx)
	}

	fmt.Printf("Select the latitude  index: ")
	fmt.Scanf("%d\n", &select_lat_idx)
	for select_lat_idx < 1 || select_lat_idx > df.row {
		fmt.Printf("%s\n",
			color.RedString(
				"Index is over the range, "+
					"please to reselect again..."))
		fmt.Printf("Select the latitude  index: ")
		fmt.Scanf("%d\n", &select_lat_idx)
	}

	fmt.Printf("Select the altitude  index: ")
	fmt.Scanf("%d\n", &select_alt_idx)
	for select_alt_idx < 1 || select_alt_idx > df.row {
		fmt.Printf("%s\n",
			color.RedString(
				"Index is over the range, "+
					"please to reselect again..."))
		fmt.Printf("Select the altitude  index: ")
		fmt.Scanf("%d\n", &select_alt_idx)
	}

	fmt.Printf("Select output filename (default: output.kml)\n")
	fmt.Printf("(enter key -> default / custom): ")
	fmt.Scanf("%s", &output_kml)

	if output_kml == "" {
		output_kml = "output.kml"
	}

	LLA := []string{}
	for i := 1; i < df.col; i++ {
		LLA = append(
			LLA,
			fmt.Sprintf(
				"%f,%f,%f",
				df.data[df.keys[select_lon_idx-1]][i],
				df.data[df.keys[select_lat_idx-1]][i],
				df.data[df.keys[select_alt_idx-1]][i],
			),
		)
		fmt.Printf("Progress = %3.2f%%\r", (float64(i+1)/float64(df.col))*100.0)
	}
	fmt.Printf("%s\r", strings.Repeat(" ", Term_Width))

	t, _ := template.New("kml").Parse(tpl)
	f, _ := os.Create(output_kml)
	defer f.Close()
	data := struct {
		FILE     string
		LLA      string
		LLA_head string
		LLA_end  string
	}{
		FILE:     filepath.Base(csvfile),
		LLA:      strings.Join(LLA, " "),
		LLA_head: LLA[0],
		LLA_end:  LLA[len(LLA)-1],
	}
	t.Execute(f, data)

	fmt.Printf("\n")
	log.Printf("End\n")
}
