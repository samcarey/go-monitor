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
)

func main() {
    http.HandleFunc("/", homepage) // setting router rule
	http.HandleFunc("/insert", insert)  // setting router rule
    http.ListenAndServe(":8080", nil) // start server 
}

func get_image(name string, data []float64, axis []float64) image.Image {
    graph := chart.Chart{
        XAxis: chart.XAxis{
            Name:      "The XAxis",
            NameStyle: chart.StyleShow(),
            Style:     chart.StyleShow(),
        },
        YAxis: chart.YAxis{
            Name:      "The YAxis",
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

    time_vals := make([]float64, num_rows)
    for i := range time_vals {
        time_vals[i] = float64(i)
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
  
    img := get_image("TestName", cols[0], time_vals)
    width := img.Bounds().Dx()
    height := img.Bounds().Dy()

    rec_large := image.Rectangle{image.Point{0, 0}, image.Point{width*ngrid_cols, height*ngrid_rows}}
    rgba := image.NewRGBA(rec_large)

    sp  := image.Point{0, 0}
    rec := image.Rectangle{sp, sp.Add(img.Bounds().Size())}
    draw.Draw(rgba, rec, img, image.Point{0, 0}, draw.Src)

    fmt.Printf("ngrid_rows: %d\n", ngrid_rows)
    fmt.Printf("ngrid_cols: %d\n", ngrid_cols)

    col_counter := 0
    for i := 0 ; i < ngrid_rows ; i++ {
        for j := 0 ; j < ngrid_cols ; j++ {
            if col_counter == 0 {
                col_counter++
                continue
            }
            if col_counter < ncols {
                fmt.Printf("col_counter / ncols: %d / %d\n", col_counter, ncols)
                img := get_image("TestName", cols[col_counter], time_vals)
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
