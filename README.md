# Furet

Simple command line interface to encrypt/decrypt data encoded with a
[Fernet](https://github.com/fernet/spec/) encryption key.


```
$ furet -h

furet encrypts or decrypts \n delimited data with Fernet.
Usage: 
    furet [-o OUTPUT] --key KEY [INPUT]
    furet [--decrypt] --key KEY [-o OUTPUT] [INPUT]
Options:
    -e, --encrypt               Encrypt the input to the output. Default if omitted.
    -d, --decrypt               Decrypt the input to the output.
    -k, --key                   Fernet key. Accepts hexadecimal standard base64 or URL-safe base64.

INPUT defaults to standard input, and OUTPUT defaults to standard output.
```

