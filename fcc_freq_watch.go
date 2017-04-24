package main

import (
	"log"
    "database/sql"
    "fmt"
    "net/http"
    _ "github.com/mattn/go-sqlite3"
    "github.com/wcharczuk/go-chart"
    "image/png"
    "bytes"
    "strconv"
    "image"
    "image/draw"
    "math"
    //"reflect"
    "html/template"
    "encoding/json"
    "os"
)

type Configuration struct {
    Red float64
    Green string
    Blue string
    Flowrate string
    Salinity string
    Timesampled string
}

func main() {
    http.HandleFunc("/", homepage) // setting router rule
	http.HandleFunc("/insert", insert)  // setting router rule
    http.HandleFunc("/config", config)
    http.ListenAndServe(":8080", nil) // start server 
}

func config(w http.ResponseWriter, r *http.Request) {
    fmt.Println("method:", r.Method) //get request method
    if r.Method == "GET" {
        t, _ := template.ParseFiles("config.gtpl")
        t.Execute(w, nil)
    } else {
        // POST method 

        // Read config file or create default config
        config_file_name := "/home/pi/config.json"
        configuration := Configuration{}
        new_config := false
        if _, err := os.Stat(config_file_name); os.IsNotExist(err) {
            new_config = true
        }
        file, _ := os.OpenFile(config_file_name, os.O_RDWR | os.O_CREATE, 0755)
        defer file.Close()

        if new_config {
            configuration.Red = 15.0
            configuration.Green = "16"
            configuration.Blue = "17"
            configuration.Flowrate = "18"
            configuration.Salinity = "19"
            configuration.Timesampled = "20"
        }else{
            decoder := json.NewDecoder(file)
            err := decoder.Decode(&configuration)
            if err != nil {
                fmt.Println("error:", err)
            }
        }

        // Make any changes to the config
        r.ParseForm()
        if red, err := strconv.ParseFloat(r.Form.Get("red"), 64) ; err == nil {configuration.Red = red}
        if green := r.Form["green"] ; len(green[0]) > 0 {configuration.Green = green[0]}
        if blue := r.Form["blue"] ; len(blue[0]) > 0 {configuration.Blue = blue[0]}
        if flowrate := r.Form["flowrate"] ; len(flowrate[0]) > 0 {configuration.Flowrate = flowrate[0]}
        if salinity := r.Form["salinity"] ; len(salinity[0]) > 0 {configuration.Salinity = salinity[0]}
        if timesampled := r.Form["timesampled"] ; len(timesampled[0]) > 0 {configuration.Timesampled = timesampled[0]}
      
        // List config 
        fmt.Println("New Configuration:")
        fmt.Printf("red: %f\n", configuration.Red)
        fmt.Printf("green: %s\n", configuration.Green)
        fmt.Printf("blue: %s\n", configuration.Blue)
        fmt.Printf("flowrate: %s\n", configuration.Flowrate)
        fmt.Printf("salinity: %s\n", configuration.Salinity)
        fmt.Printf("timesampled: %s\n", configuration.Timesampled)
        
        // Write new config file
        encoded, err := json.Marshal(configuration)
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
        file.Close()
        os.Remove(config_file_name)
        file, _ = os.OpenFile(config_file_name, os.O_RDWR | os.O_CREATE, 0755)
        file.Write(encoded)
        file.Close()
    }
}

func homepage(w http.ResponseWriter, r *http.Request) {
    
    num_rows := 300
    col_names := []string{"461.025 MHz", "461.075 MHz", "461.1 MHz", "462.155 MHz", 
                          "462.375 MHz", "462.4 MHz", "464.5 MHz", "464.55 MHz", "464.6 MHz", 
                          "464.625 MHz", "464.65 MHz", "464.725 MHz", "464.75 MHz"}
    
    ncols := len(col_names)
    cols := make([][]float64, ncols)
    entries := make([]float64, ncols*num_rows)
    for i := range cols {
        cols[i], entries = entries[:num_rows], entries[num_rows:]
    }

    ngrid_cols := 4
    ngrid_rows := int(math.Ceil(float64(ncols) / float64(ngrid_cols)))

    db, err := sql.Open("sqlite3", "/home/pi/fccFreqs.db")
    if err != nil {	fmt.Println("Failed to create the db handle") }

    rows, err := db.Query("SELECT * FROM table1")
    if err != nil {	log.Fatal(err) }
    defer rows.Close()
    
    row_counter := 0
    for rows.Next() {
		err := rows.Scan(
            &cols[0][row_counter], &cols[1][row_counter], &cols[2][row_counter], 
            &cols[3][row_counter], &cols[4][row_counter], &cols[5][row_counter], 
            &cols[6][row_counter], &cols[7][row_counter], &cols[8][row_counter], 
            &cols[9][row_counter], &cols[10][row_counter], &cols[11][row_counter], 
            &cols[12][row_counter])
		
		if err != nil {	log.Fatal(err) }
        row_counter++
	}

    max_val := 1e-12
    min_val := 1e12
    for i:= 0 ; i < len(cols) ; i++ {
        for j := 0 ; j < len(cols[0]) ; j++ {
            if cols[i][j] > max_val {
                max_val = cols[i][j]
            }
            if cols[i][j] < min_val {
                min_val = cols[i][j]
            }
        }
    }

    num_counts := 256
    hist_vals := make([]float64, num_counts)
    hist_val_inc := (max_val - min_val) / float64(num_counts)
    hist_val := min_val
    for i := range hist_vals {
        hist_vals[i] = hist_val
        hist_val += hist_val_inc
    }

 
    histogram := get_histogram(cols[0], min_val, max_val, num_counts)
    img := get_image("TestName", histogram, hist_vals, col_names[0])

    width := img.Bounds().Dx()
    height := img.Bounds().Dy()

    rec_large := image.Rectangle{image.Point{0, 0}, image.Point{width*ngrid_cols, height*ngrid_rows}}
    rgba := image.NewRGBA(rec_large)

    sp  := image.Point{0, 0}
    rec := image.Rectangle{sp, sp.Add(img.Bounds().Size())}
    draw.Draw(rgba, rec, img, image.Point{0, 0}, draw.Src)

    //fmt.Printf("ngrid_rows: %d\n", ngrid_rows)
    //fmt.Printf("ngrid_cols: %d\n", ngrid_cols)

    col_counter := 0
    for i := 0 ; i < ngrid_rows ; i++ {
        for j := 0 ; j < ngrid_cols ; j++ {
            if col_counter == 0 {
                col_counter++
                continue
            }
            if col_counter < ncols {
                fmt.Printf("col_counter / ncols: %d / %d\n", col_counter, ncols)
                histogram := get_histogram(cols[col_counter], min_val, max_val, num_counts)
                img := get_image("TestName", histogram, hist_vals, col_names[col_counter])
                col_counter++
                sp  := image.Point{j*width, i*height}
                rec := image.Rectangle{sp, sp.Add(img.Bounds().Size())}
                draw.Draw(rgba, rec, img, image.Point{0, 0}, draw.Src)
            }
        }
    }


    buffer := new(bytes.Buffer)
    if err := png.Encode(buffer, rgba); err != nil {
        log.Println("unable to encode image.")
    }

    w.Header().Set("Content-Type", "image/png")
    w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
    if _, err := w.Write(buffer.Bytes()); err != nil {
        log.Println("unable to write image.")
    }
}

func insert(w http.ResponseWriter, r *http.Request) {

	//fmt.Fprintf(w, "GET params were: %s", r.URL.Query());

	var rowCount int64 = 0

	// Pull values from URL into local variables
	paramRed := r.URL.Query().Get("red")
	paramGreen := r.URL.Query().Get("green")
	paramBlue := r.URL.Query().Get("blue")
	paramFlowrate := r.URL.Query().Get("flowrate")
	paramSalinity := r.URL.Query().Get("salinity")
	paramTimeSampled := r.URL.Query().Get("timesampled")
  	
  	if paramRed != "" && paramGreen != "" && paramBlue != "" && 
  	   paramFlowrate != "" && paramSalinity != "" && paramTimeSampled != "" {

    	
		db, err := sql.Open("sqlite3", "./database.db")

	    if err != nil {	fmt.Println("Failed to create the db handle") }

	    stmt, err := db.Prepare("INSERT INTO Samples( red, green, blue, flowrate, salinity, timesampled ) VALUES(?,?,?,?,?,?)")

		if err != nil {
			log.Fatal(err)
		}
		res, err := stmt.Exec(paramRed, paramGreen, paramBlue, paramFlowrate, paramSalinity, paramTimeSampled)
		if err != nil {
			log.Fatal(err)
		}
		lastId, err := res.LastInsertId()
		if err != nil {
			log.Fatal(err)
		}
		rowCount, err = res.RowsAffected()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("ID = %d, affected = %d\n", lastId, rowCount)
    }

    if rowCount == 1 {
    	fmt.Fprintf(w, "success")
	}else{
		fmt.Fprintf(w, "failure")
	}
}

func get_image(name string, data []float64, axis []float64, xaxis_label string) image.Image {
    graph := chart.Chart{
        XAxis: chart.XAxis{
            Name:      xaxis_label,
            NameStyle: chart.StyleShow(),
            Style:     chart.StyleShow(),
        },
        YAxis: chart.YAxis{
            Name:      "Count",
            NameStyle: chart.StyleShow(),
            Style:     chart.StyleShow(),
        },
        Series: []chart.Series{
            chart.ContinuousSeries{
                Style: chart.Style{
                    Show:        true,
                    StrokeColor: chart.GetDefaultColor(0).WithAlpha(64),
                    FillColor:   chart.GetDefaultColor(0).WithAlpha(64),
                },
                XValues: axis,
                YValues: data,
            },
        },
    }

    collector := &chart.ImageWriter{}
    graph.Render(chart.PNG, collector)
    img, err := collector.Image()
    _ = err
    return img
}

func get_histogram(data []float64, min float64, max float64, num_counts int) []float64 {
    counts := make([]float64, num_counts)
    a := max
    b := min
    c := float64(num_counts-1)
    d := float64(0)
    slope := (c-d)/(a-b)
    intercept := (d*a-b*c)/(a-b)
    for i := 0 ; i < len(data) ; i++ {
        counts[int(data[i]*slope + intercept + 0.5)]++
    }
    return counts
}

