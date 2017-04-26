package main

import (
	"log"
    "database/sql"
    "fmt"
    "net/http"
    _ "github.com/mattn/go-sqlite3"
    "github.com/wcharczuk/go-chart"
    "github.com/wcharczuk/go-chart/drawing"
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
    "syscall"
    "os/exec"
    "strings"
)

var config_file_name = "/home/pi/config.json"

func main() {
    http.HandleFunc("/", homepage) // setting router rule
	http.HandleFunc("/plot", plot)  // setting router rule
    http.HandleFunc("/config", config)
    http.HandleFunc("/shutdown", shutdown)
    http.ListenAndServe(":8080", nil) // start server 
}

func homepage(w http.ResponseWriter, r *http.Request) {
    t, _ := template.ParseFiles("home.gtpl")
    t.Execute(w, nil)
}

func shutdown(w http.ResponseWriter, r *http.Request) {
    t, _ := template.ParseFiles("shutdown.gtpl")
    t.Execute(w, nil)
    binary, lookErr := exec.LookPath("shutdown")
    if lookErr != nil {
        panic(lookErr)
    }
    args := []string{"shutdown", "-h", "now"}
    env := os.Environ()
    execErr := syscall.Exec(binary, args, env)
    if execErr != nil {
        panic(execErr)
    }
}

func config(w http.ResponseWriter, r *http.Request) {
    fmt.Println("method:", r.Method) //get request method
    if r.Method == "GET" {
        t, _ := template.ParseFiles("config.gtpl")
        t.Execute(w, nil)
    } else {
        // POST method 
        config := get_config()

        // Make any changes to the config
        r.ParseForm()
        if samp_rate, err := strconv.ParseFloat(r.Form.Get("samp_rate"), 64) ; err == nil {
            config.Samp_rate = samp_rate
        }
        if nrows, err := strconv.ParseInt(r.Form.Get("nrows"), 10, 64) ; err == nil {
            config.Nrows = int(nrows)
        }
        if database_name := r.Form["database_name"] ; len(database_name[0]) > 0 {
            config.Database_name = database_name[0]
        }
        if autoscale, err := strconv.ParseBool(r.Form.Get("autoscale")) ; err == nil {
            config.Autoscale = autoscale
        }else{
            if autoscale_int, err := strconv.ParseInt(r.Form.Get("autoscale"), 10, 64) ; err == nil {
                if autoscale_int == 0 {
                    config.Autoscale = false
                }else{
                    if autoscale_int == 1 {
                        config.Autoscale = true
                    }
                }
            }
        }
        if min_val, err := strconv.ParseFloat(r.Form.Get("min_val"), 64) ; err == nil {
            config.Min_val = min_val
        }
        if max_val, err := strconv.ParseFloat(r.Form.Get("max_val"), 64) ; err == nil {
            config.Max_val = max_val
        }
        if nhist_points, err := strconv.ParseInt(r.Form.Get("nhist_points"), 10, 64) ; err == nil {
            config.Nhist_points = int(nhist_points)
        }

        // Write new config file
        encoded, err := json.Marshal(config)
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
        os.Remove(config_file_name)
        file, _ := os.OpenFile(config_file_name, os.O_RDWR | os.O_CREATE, 0755)
        file.Write(encoded)
        file.Close()

        //fmt.Fprintf(w, "success")
        // List config on server
        fmt.Printf("New Configuration:\n")
        fmt.Printf("Sample Rate (rows/sec): %f\n", config.Samp_rate)
        fmt.Printf("Num Rows: %d (%f hours)\n", config.Nrows,(float64(config.Nrows)/(config.Samp_rate*3600.0)))
        fmt.Printf("Database Name: %s\n", config.Database_name)
        fmt.Printf("Autoscale: %t\n", config.Autoscale)
        fmt.Printf("Min Val (autoscale off): %f\n", config.Min_val)
        fmt.Printf("Max Val (autoscale off): %f\n", config.Max_val)
        fmt.Printf("Num Histogram Points: %d\n", config.Nhist_points)
        
        // List config on page
        fmt.Fprintf(w, "New Configuration:\n")
        fmt.Fprintf(w, "Sample Rate (rows/sec): %f\n", config.Samp_rate)
        fmt.Fprintf(w, "Num Rows: %d (%f hours)\n", config.Nrows,(float64(config.Nrows)/(config.Samp_rate*3600.0)))
        fmt.Fprintf(w, "Database Name: %s\n", config.Database_name)
        fmt.Fprintf(w, "Autoscale: %t\n", config.Autoscale)
        fmt.Fprintf(w, "Min Val (autoscale off): %f\n", config.Min_val)
        fmt.Fprintf(w, "Max Val (autoscale off): %f\n", config.Max_val)
        fmt.Fprintf(w, "Number of Histogram Points: %d\n", config.Nhist_points)
    }
}

type Configuration struct {
    Samp_rate float64
    Nrows int
    Database_name string
    Autoscale bool
    Min_val float64
    Max_val float64
    Nhist_points int
}

func plot(w http.ResponseWriter, r *http.Request) {
    config := get_config()

    col_names := []string{"461.025 MHz", "461.075 MHz", "461.1 MHz", "462.155 MHz", 
                          "462.375 MHz", "462.4 MHz", "464.5 MHz", "464.55 MHz", "464.6 MHz", 
                          "464.625 MHz", "464.65 MHz", "464.725 MHz", "464.75 MHz"}

    for i := 0 ; i < len(col_names) ; i++ {
        col_names[i] = strings.Join([]string{"Power [dB] at", col_names[i]}," ")
    }
    db, err := sql.Open("sqlite3", config.Database_name)
    if err != nil {	fmt.Println("Failed to create the db handle") }

    rows, err := db.Query(strings.Join([]string{"SELECT * FROM table1 ASC limit", strconv.Itoa(config.Nrows)}," "))
    //rows, err := db.Query("SELECT * FROM table1")
    if err != nil {	log.Fatal(err) }
    nrows := 0
    for rows.Next() {nrows++}
    rows.Close()
    rows, err = db.Query(strings.Join([]string{"SELECT * FROM table1 ASC limit", strconv.Itoa(nrows)}," "))
    //rows, err := db.Query("SELECT * FROM table1")
    if err != nil {	log.Fatal(err) }

    fmt.Printf("Retrieved / Requested: %d / %d\n", nrows, config.Nrows)

    ncols := len(col_names)
    cols := make([][]float64, ncols)
    entries := make([]float64, ncols*nrows)
    for i := range cols {
        cols[i], entries = entries[:nrows], entries[nrows:]
    }

    ngrid_cols := 4
    ngrid_rows := int(math.Ceil(float64(ncols) / float64(ngrid_cols)))
    
    for i := 0 ; i < nrows && rows.Next() ; i++ {
		err := rows.Scan(
            &cols[0][i], &cols[1][i], &cols[2][i], &cols[3][i], &cols[4][i], &cols[5][i], &cols[6][i], 
            &cols[7][i], &cols[8][i], &cols[9][i], &cols[10][i], &cols[11][i], &cols[12][i])
		
		if err != nil {	log.Fatal(err) }
	}


    min_val := config.Min_val
    max_val := config.Max_val
    if config.Autoscale {
        max_val = 1e-12
        min_val = 1e12
        for i:= 0 ; i < len(cols) ; i++ {
            for j := 0 ; j < len(cols[0]) ; j++ {
                //fmt.Printf("min: %f, max: %f, this: %f\n", min_val, max_val, cols[i][j])
                if cols[i][j] > max_val {
                    max_val = cols[i][j]
                }
                if cols[i][j] < min_val {
                    min_val = cols[i][j]
                }
            }
        }
    }

    hist_vals := make([]float64, config.Nhist_points)
    hist_val_inc := (max_val - min_val) / float64(config.Nhist_points)
    hist_val := min_val
    for i := range hist_vals {
        hist_vals[i] = hist_val
        hist_val += hist_val_inc
    }

 
    histogram := get_histogram(cols[0], min_val, max_val, config.Nhist_points)
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
                histogram := get_histogram(cols[col_counter], min_val, max_val, config.Nhist_points)
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

func get_image(name string, data []float64, axis []float64, xaxis_label string) image.Image {
    graph := chart.Chart{
        XAxis: chart.XAxis{
            Name:      xaxis_label,
            NameStyle: chart.StyleShow(),
            Style:     chart.StyleShow(),
        },
        YAxis: chart.YAxis{
            Name:      "log10(Count)",
            NameStyle: chart.StyleShow(),
            Style:     chart.StyleShow(),
        },
        Series: []chart.Series{
            chart.ContinuousSeries{
                Style: chart.Style{
                    Show:        true,
                    StrokeColor: drawing.Color{R: 0, G: 0, B: 0, A: 255},
                    FillColor:   drawing.Color{R: 0, G: 0, B: 0, A: 255},
                    //StrokeColor: chart.GetDefaultColor(0).WithAlpha(64),
                    //FillColor:   chart.GetDefaultColor(0).WithAlpha(64),
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
        if data[i] >= min && data[i] <= max {
            counts[int(data[i]*slope + intercept + 0.5)]++
        }
    }
    for i := 0 ; i < len(counts) ; i++ {
        if counts[i] != 0 {
            counts[i] = math.Log10(counts[i])
        }
    }
    return counts
}

func get_config() Configuration {
    // Read config file or create default config
    config := Configuration{}
    new_config := false
    if _, err := os.Stat(config_file_name); os.IsNotExist(err) {
        new_config = true
    }
    file, _ := os.OpenFile(config_file_name, os.O_RDWR | os.O_CREATE, 0755)
    defer file.Close()

    if new_config {
        config.Samp_rate = 4
        config.Nrows = int(config.Samp_rate*60*60)
        config.Database_name = "/home/pi/fccFreqs.db"
        config.Autoscale = true
        config.Min_val = -100.0
        config.Max_val = 0.0
        config.Nhist_points = 256
    }else{
        decoder := json.NewDecoder(file)
        err := decoder.Decode(&config)
        if err != nil {
            fmt.Println("error:", err)
        }
    }
    return config
}

