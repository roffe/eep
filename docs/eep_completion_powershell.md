## eep completion powershell

Generate the autocompletion script for powershell

### Synopsis

Generate the autocompletion script for powershell.

To load completions in your current shell session:

	eep completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.


```
eep completion powershell [flags]
```

### Options

```
  -h, --help              help for powershell
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

