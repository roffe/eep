## eep completion zsh

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions for every new session, execute once:

#### Linux:

	eep completion zsh > "${fpath[1]}/_eep"

#### macOS:

	eep completion zsh > /usr/local/share/zsh/site-functions/_eep

You will need to start a new shell for this setup to take effect.


```
eep completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
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

