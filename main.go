package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/fernet/fernet-go"
	"golang.org/x/term"
)

const usage = `
furet encrypts or decrypts \n delimited data with Fernet.
Usage: 
    furet [-o OUTPUT] --key KEY [INPUT]
    furet [--decrypt] --key KEY [-o OUTPUT] [INPUT]
Options:
    -e, --encrypt               Encrypt the input to the output. Default if omitted.
    -d, --decrypt               Decrypt the input to the output.
    -k, --key                   Fernet key. Accepts hexadecimal standard base64 or URL-safe base64.

INPUT defaults to standard input, and OUTPUT defaults to standard output.`

func main() {
	log.SetPrefix("furet:")
	log.SetFlags(0)
	flag.Usage = func() { fmt.Fprintf(os.Stderr, "%s\n", usage) }

	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(1)
	}

	var (
		outFlag                  string
		decryptFlag, encryptFlag bool
		keyFlag                  string
	)

	flag.BoolVar(&decryptFlag, "d", false, "decrypt the input")
	flag.BoolVar(&decryptFlag, "decrypt", false, "decrypt the input")
	flag.BoolVar(&encryptFlag, "e", false, "encrypt the input")
	flag.BoolVar(&encryptFlag, "encrypt", false, "encrypt the input")
	flag.StringVar(&outFlag, "o", "", "output to `FILE` (default stdout)")
	flag.StringVar(&outFlag, "output", "", "output to `FILE` (default stdout)")
	flag.StringVar(&keyFlag, "k", "", "fernet key")
	flag.StringVar(&keyFlag, "key", "", "fernet key")
	flag.Parse()

	if keyFlag == "" {
		errorf("--key flag is mandatory")
	}

	var (
		in         io.Reader = os.Stdin
		out        io.Writer = os.Stdout
		stdinInUse bool
	)
	if name := flag.Arg(0); name != "" && name != "-" {
		f, err := os.Open(name)
		if err != nil {
			errorf("failed to open input file %q: %v", name, err)
		}
		defer f.Close()
		in = f
	} else {
		stdinInUse = true
	}
	if name := outFlag; name != "" && name != "-" {
		f := newLazyOpener(name)
		defer func() {
			if err := f.Close(); err != nil {
				errorf("failed to close output file %q: %v", name, err)
			}
		}()
		out = f
	} else if term.IsTerminal(int(os.Stdout.Fd())) {
		if name != "-" {
			if decryptFlag {
				// TODO: buffer the output and check it's printable.
			}
		}
		if in == os.Stdin && term.IsTerminal(int(os.Stdin.Fd())) {
			// If the input comes from a TTY and output will go to a TTY,
			// buffer it up so it doesn't get in the way of typing the input.
			buf := &bytes.Buffer{}
			defer func() { io.Copy(os.Stdout, buf) }()
			out = buf
		}
	}

	_ = stdinInUse

	key, err := fernet.DecodeKey(keyFlag)
	if err != nil {
		errorf("can't decode Fernet key: %s", err)
	}

	switch {
	case decryptFlag:
		decrypt(key, in, out)
	default:
		encrypt(key, in, out)
	}
}

func decrypt(key *fernet.Key, in io.Reader, out io.Writer) {
	scan := bufio.NewScanner(in)
	for scan.Scan() {
		msg := fernet.VerifyAndDecrypt(scan.Bytes(), 0, []*fernet.Key{key})
		if msg == nil {
			errorf("can't decrypt input %q", string(scan.Bytes()))
		}
		if _, err := fmt.Fprintf(out, "%s\n", string(msg)); err != nil {
			errorf("error writing to output: %s", err)
		}
	}

	if scan.Err() != nil {
		errorf("error reading from input: %s", scan.Err())
	}
}

func encrypt(key *fernet.Key, in io.Reader, out io.Writer) {
	scan := bufio.NewScanner(in)
	for scan.Scan() {
		msg, err := fernet.EncryptAndSign(scan.Bytes(), key)
		if err == nil {
			errorf("can't encrypt input %q: %s", string(scan.Bytes()), err)
		}
		if _, err := fmt.Fprintf(out, "%s\n", string(msg)); err != nil {
			errorf("error writing to output: %s", err)
		}
	}

	if scan.Err() != nil {
		errorf("error reading from input: %s", scan.Err())
	}
}

type lazyOpener struct {
	name string
	f    *os.File
	err  error
}

func newLazyOpener(name string) io.WriteCloser {
	return &lazyOpener{name: name}
}

func (l *lazyOpener) Write(p []byte) (n int, err error) {
	if l.f == nil && l.err == nil {
		l.f, l.err = os.Create(l.name)
	}
	if l.err != nil {
		return 0, l.err
	}
	return l.f.Write(p)
}

func (l *lazyOpener) Close() error {
	if l.f != nil {
		return l.f.Close()
	}
	return nil
}

func errorf(format string, v ...interface{}) {
	log.Printf("error: "+format, v...)
	log.Fatalf("report unexpected or unhelpful errors at https://github.com/arl/furet")
}

func warningf(format string, v ...interface{}) {
	log.Printf("warning: "+format, v...)
}

func errorWithHint(error string, hints ...string) {
	log.Printf("error: %s", error)
	for _, hint := range hints {
		log.Printf("hint: %s", hint)
	}
	log.Fatalf("report unexpected or unhelpful errors at https://github.com/arl/furet")
}
