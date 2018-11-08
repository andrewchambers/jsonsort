package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/valyala/fastjson"
	"golang.org/x/sync/errgroup"
)

var (
	sortMethod = flag.String("method", "", "For gnusort, valid methods are general-numeric, human-numeric, month, numeric, random or version sort.")
	command    = flag.String("command", "sort", "Sort command binary (usually should be gnu sort).")
	ignoreCase = flag.Bool("ignore-case", false, "Ignore case in key.")
	unique     = flag.Bool("unique", false, " Don't print json objects with duplicate keys, unique keys only.")
	debug      = flag.Bool("debug", false, "Debug output to stderr.")
)

const DELIM = 0x02

func main() {
	flag.Parse()

	keyPath := flag.Args()

	if len(keyPath) == 0 {
		fmt.Fprintf(os.Stderr, "You must specify a path to the sort key.\n")
		os.Exit(1)
	}

	eg, ctx := errgroup.WithContext(context.Background())

	args := []string{"--compress-program=gzip", "-t", "\x02"}

	if *sortMethod != "" {
		args = append(args, "--sort="+*sortMethod)
	}

	if *unique {
		args = append(args, "-u")
	}

	if *ignoreCase {
		args = append(args, "--ignore-case")
	}

	if *debug {
		fmt.Fprintf(os.Stderr, "Sort command: '%#v' args=%#v\n", *command, args)
	}

	sortCmd := exec.CommandContext(ctx, *command, args...)

	sortCmd.Stderr = os.Stderr

	sortIn, err := sortCmd.StdinPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting sort input pipe: %s\n", err)
		os.Exit(1)
	}

	sortOut, err := sortCmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting sort output pipe: %s\n", err)
		os.Exit(1)
	}

	eg.Go(func() error {
		jsonParser := fastjson.Parser{}
		brdr := bufio.NewReader(os.Stdin)
		var sortBuf bytes.Buffer

		for {
			ln, lnError := brdr.ReadBytes('\n')

			if len(ln) == 0 {
				if lnError != nil {
					if lnError == io.EOF {
						err := sortIn.Close()
						if err != nil {
							return err
						}
						return nil
					}
					return err
				}
				panic("expected error for empty read")
			}

			jsonv, err := jsonParser.ParseBytes(ln)
			if err != nil {
				return err
			}

			sortBuf.Reset()

			key := jsonv.Get(keyPath...)
			if key != nil {
				marshaledKey := key.MarshalTo(nil)
				if err != nil {
					return err
				}
				if key.Type() == fastjson.TypeString {
					marshaledKey = marshaledKey[1 : len(marshaledKey)-1]
				}
				if *debug {
					fmt.Fprintf(os.Stderr, "Sort key: '%v'\n", string(marshaledKey))
				}
				_, err = sortBuf.Write(marshaledKey)
				if err != nil {
					return err
				}
			} else {
				_, _ = fmt.Fprintf(os.Stderr, "Warning, line missing sort key\n")
			}

			_, err = sortBuf.Write([]byte{DELIM})
			if err != nil {
				return err
			}
			_, err = sortBuf.Write(ln)
			if err != nil {
				return err
			}

			_, err = sortIn.Write(sortBuf.Bytes())
			if err != nil {
				return err
			}
		}
	})

	eg.Go(func() error {
		brdr := bufio.NewReader(sortOut)

		for {
			ln, lnError := brdr.ReadBytes('\n')

			if len(ln) == 0 {
				if lnError != nil {
					if lnError == io.EOF {
						err := sortIn.Close()
						if err != nil {
							return err
						}
						return nil
					}
					return err
				}
				panic("expected error for empty read")
			}

			sepIdx := bytes.Index(ln, []byte{DELIM})
			if sepIdx == -1 {
				return errors.New("Sort returned line without x02 separator!")
			}

			_, err = os.Stdout.Write(ln[sepIdx+1:])
			if err != nil {
				return err
			}
		}
	})

	eg.Go(func() error {
		err = sortCmd.Start()
		if err != nil {
			return err
		}
		return sortCmd.Wait()
	})

	err = eg.Wait()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during sorting: %s\n", err)
		os.Exit(1)
	}

}
