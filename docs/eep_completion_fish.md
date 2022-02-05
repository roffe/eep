## eep completion fish

Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	eep completion fish | source

To load completions for every new session, execute once:

	eep completion fish > ~/.config/fish/completions/eep.fish

You will need to start a new shell for this setup to take effect.


```
eep completion fish [flags]
```

### Options

```
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
  -c, --chip uint8     chip type (default 66)
  -e, --erase          erase before write (default false
  -o, --org uint8      chip org (default 8)
  -p, --port string    com port (default "COM8")
  -s, --size uint16    chip size (default 512)
  -x, --xor bytesHex   xor output (default 00)
```

### SEE ALSO

* [eep completion](eep_completion.md)	 - Generate the autocompletion script for the specified shell

