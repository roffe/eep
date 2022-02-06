#include <Arduino.h>
#include <EEPROM.h>
#include "M93Cx6.h"

#define USE_EEPROM true
#define PWR_PIN 9
#define CS_PIN 10
#define SK_PIN 7
#define DI_PIN 11
#define DO_PIN 12
#define ORG_PIN 8

M93Cx6 ep = M93Cx6(PWR_PIN, CS_PIN, SK_PIN, DO_PIN, DI_PIN, ORG_PIN);

uint8_t incomingByte;

int chipAddr = 0x08;
int orgAddr = 0x10;
int sizeAddr = 0x20;

static uint8_t cfgChip;
static uint16_t size;
static uint8_t org;

void setup()
{
    Serial.begin(57600);

    if (USE_EEPROM)
    {
        loadSettings();
    }

    pinMode(LED_BUILTIN, OUTPUT); // LED
    while (!Serial)
    {
        ; // wait for serial port to connect. Needed for native USB
    }
    Serial.print("\f");
}

void loop()
{
    handleSerial();
}

void loadSettings()
{
    uint8_t check;
    EEPROM.get(0x00, check);
    if (check == 0x20)
    {
        cfgChip = EEPROM.read(chipAddr);
        EEPROM.get(sizeAddr, size);
        org = EEPROM.read(orgAddr);
        ep.setChip(cfgChip);
        ep.setOrg(org);
    }
}

static int buffLength = 0; // number of characters currently in the buffer

void handleSerial()
{
    const int BUFF_SIZE = 32;          // make it big enough to hold your longest command
    static char buffer[BUFF_SIZE + 1]; // +1 allows space for the null terminator

    if (Serial.available())
    {
        char c = Serial.read();
        // Serial.println(c, HEX);

        if ((c == '\r') || (c == '\n'))
        {
            // end-of-line received
            if (buffLength > 0)
            {
                handleReceivedMessage(buffer);
            }
            buffLength = 0;
        }
        else if (c == 127) // handle backspace during command parsing mode
        {
            buffLength--;
        }
        else
        {
            if (buffLength < BUFF_SIZE)
            {
                buffer[buffLength++] = c; // append the received character to the array
                buffer[buffLength] = 0;   // append the null terminator
            }
            else
            {
                Serial.write('\a');
            }
        }
    }
}

char *cmd;

void handleReceivedMessage(char *msg)
{
    switch (msg[0])
    {
    case 'h':
        help();
        return;
    case 's':
        parseMsg(msg);
        Serial.println();
        return;
    case 'w':
    case 'r':
    case 'e':
    case 'p':
        break;
    case '?':
        settings();
        return;
    default:
        Serial.println("invalid command");
        help();
        return;
    }

    if (buffLength > 4) // long command with set options inline
    {
        parseMsg(msg);
    }

    ep.powerUp();
    delayMicroseconds(100);

    switch (msg[0])
    {
    case 'r':
        read();
        break;
    case 'w':
        write();
        break;
    case 'p':
        printBin();
        break;
    case 'e':
        erase();
        break;
    }
}

void parseMsg(char *msg)
{
    // Serial.println("parse");
    char *pos;
    cmd = strtok(msg, ",");

    pos = strtok(NULL, ",");
    cfgChip = atoi(pos);

    pos = strtok(NULL, ",");
    size = atoi(pos);

    pos = strtok(NULL, ",");
    org = atoi(pos);

    if (size == 0)
    {
        Serial.println("\ainvalid size");
        return;
    }
    if (cfgChip == 0)
    {
        Serial.println("\ainvalid chip");
        return;
    }
    if (org == 0)
    {
        Serial.println("\ainvalid org");
        return;
    }

    if (!setChip(cfgChip))
    {
        return;
    }

    if (!setOrg(org))
    {
        return;
    }

    if (USE_EEPROM)
    {
        EEPROM.put(sizeAddr, size);
        EEPROM.put(0x00, 0x20); // if this is not 0x20 settings will not be loaded from eeprom
    }
}

bool setChip(uint8_t chip)
{
    bool ok = false;
    switch (chip)
    {
    case 46:
        ep.setChip(M93C46);
        ok = true;
        break;
    case 56:
        ep.setChip(M93C56);
        ok = true;
        break;
    case 66:
        ep.setChip(M93C66);
        ok = true;
        break;
    case 76:
        ep.setChip(M93C76);
        ok = true;
        break;
    case 86:
        ep.setChip(M93C86);
        ok = true;
        break;
    default:
        Serial.println("\ainvalid CHIP");
        return false;
    }
    if (!ok)
    {
        return false;
    }
    EEPROM.put(chipAddr, chip);
    return true;
}

bool setOrg(uint8_t org)
{
    bool ok = false;
    switch (org)
    {
    case 8:
        ep.setOrg(ORG_8);
        ok = true;
        break;
    case 16:
        ep.setOrg(ORG_16);
        ok = true;
        break;
    default:
        Serial.println("\ainvalid ORG");
        return false;
    }
    if (!ok)
    {
        return false;
    }

    EEPROM.put(orgAddr, org);
    return true;
}

void help()
{
    Serial.println("--- eep ---");
    Serial.println("s,<chip>,<size>,<org> - Set eeprom options");
    Serial.println("? - Print current settings");
    Serial.println("r - Read eeprom");
    Serial.println("w - Initiate write mode");
    Serial.println("e - Erase eeprom");
    Serial.println("p - Hex print eeprom content");
    Serial.println("h - This help");
}

void settings()
{
    Serial.println("--- settings ---");
    Serial.print("chip: ");
    Serial.println(cfgChip);
    Serial.print("size: ");
    Serial.println(size);
    Serial.print("org: ");
    Serial.println(org);
}

void read()
{
    uint8_t c;
    ledOn();
    for (size_t i = 0; i < size; i++)
    {
        c = ep.read(i);
        Serial.write(c);
    }
    Serial.print("\r\n");
    ledOff();
    ep.powerDown();
}

void write()
{
    uint8_t c;
    unsigned long lastData;
    lastData = millis();

    Serial.write('\f');
    ledOn();
    ep.writeEnable();
    for (size_t i = 0; i < size; i++)
    {
        while (Serial.available() == 0)
        {
            if ((millis() - lastData) > 2000)
            {
                ep.writeDisable();
                ep.powerDown();
                ledOff();
                Serial.println("\adata read timeout");
                return;
            }
        }

        c = Serial.read();
        ep.write(i, c);
        lastData = millis();
    }
    ep.writeDisable();
    ep.powerDown();
    ledOff();
    Serial.print("\fwrite done\r\n");
}

void erase()
{
    ledOn();
    ep.powerUp();
    delayMicroseconds(50);
    ep.writeEnable();
    delayMicroseconds(50);
    ep.eraseAll();
    delayMicroseconds(50);
    ep.writeDisable();
    delayMicroseconds(50);
    ep.powerDown();
    ledOff();
    Serial.write('\f');
}

void printBin()
{
    uint8_t pos = 0;
    uint8_t c;
    char buf[3];
    ledOn();
    Serial.println("--- Hex dump ---");
    for (size_t i = 0; i < size; i++)
    {
        c = ep.read(i);
        sprintf(buf, "%02X ", c);
        Serial.print(buf);
        pos++;
        if (pos == 25)
        {
            Serial.println();
            pos = 0;
        }
    }
    Serial.print("\r\n");
    ledOff();
    ep.powerDown();
}

void ledOn()
{
    digitalWrite(LED_BUILTIN, HIGH);
}

void ledOff()
{
    digitalWrite(LED_BUILTIN, LOW);
}