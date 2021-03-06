# Arduino 93Cx6 EEPROM programmer with go cross-platform CLI-tool

![unit](unit.jpg)
(image has external led on pin13 & push button from other project hooked up)

    /!\ ORG_16 is not working at the moment, _only__ ORG_8

## Hardware

* Arduino UNO
* breadboard
* SOP8 clamp + flat cable
* breadboard jumper wires


## Schematics

                          _____
     Chip Select (cs)  --|1   8|--  (pwr) Vcc
    Serial Clock (sk)  --|2   7|--
         Data In (di)  --|3   6|--  (org) Organization Select
        Data Out (do)  --|4   5|--  (gnd) Vss/Ground
                          ¯¯¯¯¯
     Arduino Connection:
       Vcc (pwr)         - Pin 9
       Vss (gnd)         - GND
       Chip Select (cs)  - Pin 10
       Serial Clock (sk) - Pin 7
       Data In (di)      - Pin 11
       Data Out (do)     - Pin 12
       Org Select (org)  - Pin 8

The red cable on the SOP8 clamp indicates pin 1 and should be orinted to the corner marked with a small dot on the EEPROM

       8   7   6   5
     _ | _ | _ | _ | _
    |                 |
    |                 |
    | O               |
     ¯ | ¯ | ¯ | ¯ | ¯
       1   2   3   4


## [Docs](docs/eep.md)