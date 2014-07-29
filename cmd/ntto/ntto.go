package main

import (
	"flag"
	"fmt"
	"github.com/miku/ntto"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"runtime/pprof"
)

func main() {

	executive := "replace"
	_, err := exec.LookPath("replace")
	if err != nil {
		executive = "perl"
	}

	_, err = exec.LookPath("perl")
	if err != nil {
		log.Fatalln("This program requires perl or replace.")
		os.Exit(1)
	}

	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	version := flag.Bool("v", false, "prints current version and exits")
	dumpRules := flag.Bool("d", false, "dump rules and exit")
	dumpCommand := flag.Bool("c", false, "dump constructed sed command and exit")
	rulesFile := flag.String("r", "", "path to rules file, use built-in if none given")
	outFile := flag.String("o", "", "output file to write result to")
	nullValue := flag.String("n", "<NULL>", "string to indicate empty string replacement")
	workers := flag.Int("w", runtime.NumCPU(), "number of sed processes")

	flag.Parse()

	var PrintUsage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] FILE\n", os.Args[0])
		flag.PrintDefaults()
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatalln(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if *version {
		fmt.Println(ntto.AppVersion)
		os.Exit(0)
	}

	var rules []ntto.Rule

	if *rulesFile == "" {
		rules, err = ntto.ParseRules(ntto.DefaultRules)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		b, err := ioutil.ReadFile(*rulesFile)
		if err != nil {
			log.Fatalln(err)
		}
		rules, err = ntto.ParseRules(string(b))
		if err != nil {
			log.Fatalln(err)
		}
	}

	if *dumpRules {
		fmt.Println(ntto.DumpRules(rules))
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		PrintUsage()
		os.Exit(1)
	}

	filename := flag.Args()[0]
	var output string

	if *outFile == "" {
		tmp, err := ioutil.TempFile("", "ntto-")
		output = tmp.Name()
		fmt.Fprintf(os.Stderr, "Writing to %s\n", output)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		output = *outFile
	}

	var command string
	if executive == "perl" {
		command = fmt.Sprintf("%s > %s", ntto.SedifyNull(rules, *workers, filename, *nullValue), output)
	} else {
		command = fmt.Sprintf("%s > %s", ntto.ReplacifyNull(rules, filename, *nullValue), output)
	}

	if *dumpCommand {
		fmt.Println(command)
		os.Exit(0)
	}

	_, err = exec.Command("sh", "-c", command).Output()
	if err != nil {
		log.Fatalln(err)
	}
}
