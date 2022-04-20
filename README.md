# Furet

Simple command line interface to encrypt/decrypt data encoded with a
[Fernet](https://github.com/fernet/spec/) encryption key.


```
$ furet -h

furet encrypts or decrypts with the Fernet symmetric encryption.

Usage:
 
    furet [-o OUTPUT] --key KEY [INPUT]
    furet [--decrypt] --key KEY [-o OUTPUT] [INPUT]

Options:
    -e, --encrypt     Encrypt the input to the output. Default if omitted.
    -d, --decrypt     Decrypt the input to the output.
    -k, --key         Key to use. (accepts hexadecimal standard base64 or URL-safe base64.
    -g, --generate    Generate a random key.
INPUT defaults to standard input, and OUTPUT defaults to standard output.

Examples:
    $ KEY=$(furet -g)
    $ furet --key $KEY -o file.furet file
    $ furet --key $KEY -o file.furet < file
    $ furet --decrypt -k $KEY -o file file.furet	
    $ furet --decrypt -k $KEY < file.furet > file`
```

