package main

import (
        "fmt"
        "time"
        "net/http"
        "path/filepath"
        "sort"
        "strconv"
	"regexp"
	"errors"
        "os"
        "io/ioutil"
        "os/exec"
)

func main() {
        http.HandleFunc("/", runProcessHandler)
        http.HandleFunc("/log/", printLogHandler)
        http.HandleFunc("/list", runIndexHandler)
        http.ListenAndServe(":"+os.Args[1], nil)
}

// Print index

func runIndexHandler(w http.ResponseWriter, r *http.Request) {
	// GBet log files 
        logFiles, e := filepath.Glob("*/log*.txt")
        if e!=nil {
                fmt.Fprintf(w, "%v", e)
                return
        }

	// Sort them
        stats := make([]file, 0)
        for _, v := range logFiles {
                fi, e := os.Stat(v)
                if e != nil {
                        fmt.Fprintf(w, "Problem stating: %s", e)
                } else {
                        stats = append(stats, file{fi, v})
                }
        }
        sort.Sort(files(stats))

	// Print them all out
        for _, v := range stats {
                w.Header().Set("Content-type", "text/html")
                fmt.Fprintf(w, "<ul>")
                loc, _ := time.LoadLocation("Europe/London");
                fmt.Fprintf(w, "<li>%s <a href=\"log/%s\">%s<a></li>", v.ModTime().In(loc).Format("02 Jan 06 15:04"), v.FullPath, v.FullPath)
                fmt.Fprintf(w, "</ul>")
        }
}

type file struct {
        os.FileInfo
        FullPath string
}

type files []file

func (f files) Len() int {
        return len(f)
}

func (f files) Less(i, j int) bool {
        return f[i].ModTime().After(f[j].ModTime())
}

func (f files) Swap(i, j int) {
        t := f[i]
        f[i] = f[j]
        f[j] = t
}

// Run job handler

func runProcessHandler(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-type", "text/html")
        name := r.URL.Path[1:]

	// Check if exists
	if _, e := os.Stat(name); e != nil || os.IsNotExist(e) {
		fmt.Fprintf(w, "Project doesn't exist.",)
                return
	}

        // Print build details
        fmt.Fprintf(w, "Running...<br /><br />")
	// TODO We can't tell the user what their log number is yet
        fmt.Fprintf(w, "<a href='list'>index</a>");

        // Run process
        go runProcess(name)
}

func runProcess(name string) {
	num, err := findNextLogfileNumber(name)
	// TODO: The user won't know if this failure has happened
	if err != nil {
		fmt.Printf("Error: %v", err);
		return;
	}
	fmt.Println("Log number ", num)
        runProcessAndOutputLog(name, num)
}

func findNextLogfileNumber(name string) (int, error) {
	num, e := numFilesByGlob(name+"/log*")
        if e!=nil || num < 0 {
		return -1, errors.New("Can't list log files in the project directory.")
        }
        num++;
	return num, nil
}

func numFilesByGlob(dir string) (int, error) {
        files, e := filepath.Glob(dir)
        return len(files), e
}

func runProcessAndOutputLog(name string, num int) {
	fmt.Printf("STARTED BUILD SCRIPT: %v\n", name)
        // Run command
        cmd := exec.Command("bash", "./"+name+"/script.sh")

        // Get log file
        outfile, err := os.Create(name+"/log"+strconv.Itoa(num)+".txt")
        if err != nil {
		// TODO: The uesr won't know if this has happened
                fmt.Printf("%v\n", err)
                return
        }
        defer outfile.Close()
        cmd.Stdout = outfile

        // Run command
        err = cmd.Run()
	fmt.Printf("FINISHED BUILD SCRIPT: %v\n", name)
        if err != nil {
                fmt.Printf("%v\n", err)
        } else {
		runPipelineProjects(name)
	}
	return
}

func runPipelineProjects(name string) {
	pipelineProjects, err := findPiplineProjects(name)
	if err != nil {
		fmt.Printf("Error: Pipline projects: %v\n", err);
	} else if pipelineProjects != nil {
		fmt.Printf("Found pipeline projects, %v\n", pipelineProjects);
		for _, v := range pipelineProjects {
			runProcess(v)
		}
	}
}

func findPiplineProjects(name string) ([]string, error) {
	rf, err := ioutil.ReadFile(name+"/script.sh")
	if err != nil {
		return nil, err
	}
	file := string(rf)

	rp := regexp.MustCompile("# PIPELINE: (.*)")
	matches := rp.FindAllStringSubmatch(file, -1)

	m := make([]string, 0)
	if matches != nil {
		for _ ,v := range matches {
			m = append(m, v[1])
		}
		return m, nil
	}
	return nil, nil
}

// Log handler

func printLogHandler(w http.ResponseWriter, r *http.Request) {
        name := r.URL.Path[5:]
	// I'M FAIRLY SURE THIS IS A SECURITY HOLE.
	// WATCH ME FIX THIS. WATCH ME FIX THIS RIGHT NOW.
        fileBytes, e := ioutil.ReadFile(name)
        if e!=nil {
                fmt.Fprintf(w, "%v", e)
        } else {
                fmt.Fprintf(w, string(fileBytes))
        }
}
