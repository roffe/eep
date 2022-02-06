#ifndef USE_EEPROM
#define USE_EEPROM
#endif

#include <Arduino.h>

#ifdef USE_EEPROM
#include <EEPROM.h>
#endif

#include "M93Cx6.h"

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
static uint16_t cfgSize;
static uint8_t cfgOrg;

void setup()
{
    Serial.begin(57600);

#ifdef USE_EEPROM
    loadSettings();
#endif

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

#ifdef USE_EEPROM
// load settings from eeprom
void loadSettings()
{
    uint8_t check;
    EEPROM.get(0x00, check);
    if (check == 0x20)
    {
        cfgChip = EEPROM.read(chipAddr);
        EEPROM.get(sizeAddr, cfgSize);
        cfgOrg = EEPROM.read(orgAddr);
        ep.setChip(cfgChip);
        ep.setOrg(cfgOrg);
    }
}
#endif

static uint8_t buffLength = 0;     // number of characters currently in the buffer
const uint8_t BUFF_SIZE = 16;      // make it big enough to hold your longest command
static char buffer[BUFF_SIZE + 1]; // +1 allows space for the null terminator

void handleSerial()
{
    if (Serial.available())
    {
        char c = Serial.read();
        if ((c == '\r') || (c == '\n'))
        {
            // end-of-line received
            if (buffLength > 0)
            {
                handleCmd(buffer);
            }
            buffLength = 0;
        }
        else if (c == 127) // handle backspace during command input
        {

            if (buffLength > 0)
            {
                buffLength--;
            }
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

void handleCmd(char *msg)
{
    switch (msg[0])
    {
    case 'h':
        help();
        return;
    case 's':
        parse(msg);
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
        parse(msg);
    }
    ledOn();
    ep.powerUp();

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

    delayMicroseconds(50);
    ep.powerDown();
    ledOff();
}

static char *cmd;

void parse(char *msg)
{
    char *pos;
    cmd = strtok(msg, ",");

    pos = strtok(NULL, ",");
    cfgChip = atoi(pos);

    pos = strtok(NULL, ",");
    cfgSize = atoi(pos);

    pos = strtok(NULL, ",");
    cfgOrg = atoi(pos);

    if (cfgSize == 0)
    {
        Serial.println("\ainvalid size");
        return;
    }
    if (cfgChip == 0)
    {
        Serial.println("\ainvalid chip");
        return;
    }
    if (cfgOrg == 0)
    {
        Serial.println("\ainvalid org");
        return;
    }

    if (!setChip(cfgChip))
    {
        return;
    }

    if (!setOrg(cfgOrg))
    {
        return;
    }

#ifdef USE_EEPROM
    EEPROM.put(sizeAddr, cfgSize);
    EEPROM.put(0x00, 0x20); // if this is not 0x20 settings will not be loaded from eeprom
#endif
}

bool setChip(uint8_t chip)
{
    switch (chip)
    {
    case 46:
        ep.setChip(M93C46);
        break;
    case 56:
        ep.setChip(M93C56);
        break;
    case 66:
        ep.setChip(M93C66);
        break;
    case 76:
        ep.setChip(M93C76);
        break;
    case 86:
        ep.setChip(M93C86);
        break;
    default:
        Serial.println("\ainvalid CHIP");
        return false;
    }

#ifdef USE_EEPROM
    EEPROM.put(chipAddr, chip);
#endif

    return true;
}

bool setOrg(uint8_t org)
{
    switch (org)
    {
    case 8:
        ep.setOrg(ORG_8);
        break;
    case 16:
        ep.setOrg(ORG_16);
        break;
    default:
        Serial.println("\ainvalid ORG");
        return false;
    }

#ifdef USE_EEPROM
    EEPROM.put(orgAddr, org);
#endif

    return true;
}

void help()
{
    Serial.println("--- eep ---");
    Serial.println("s,<chip>,<size>,<org> - Set eeprom configuration");
    Serial.println("? - Print current configuration");
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
    Serial.println(cfgSize);
    Serial.print("org: ");
    Serial.println(cfgOrg);
}

void read()
{
    for (uint16_t i = 0; i < cfgSize; i++)
    {
        Serial.write(ep.read(i));
    }
    Serial.print("\r\n");
}

static unsigned long lastData;

void write()
{
    lastData = millis();
    Serial.write('\f');
    ep.writeEnable();
    for (uint16_t i = 0; i < cfgSize; i++)
    {
        while (Serial.available() == 0)
        {
            if ((millis() - lastData) > 2000)
            {
                ep.writeDisable();
                Serial.println("\adata read timeout");
                return;
            }
        }
        ep.write(i, Serial.read());
        lastData = millis();
    }
    ep.writeDisable();
    Serial.print("\fwrite done\r\n");
}

void erase()
{
    ep.writeEnable();
    ep.eraseAll();
    ep.writeDisable();
    Serial.write('\f');
}

static uint8_t linePos;
void printBin()
{
    linePos = 0;
    char buf[3];
    ledOn();
    Serial.println("--- Hex dump ---");
    for (uint16_t i = 0; i < cfgSize; i++)
    {
        sprintf(buf, "%02X ", ep.read(i));
        Serial.print(buf);
        linePos++;
        if (linePos == 24)
        {
            Serial.println();
            linePos = 0;
        }
    }
    Serial.println();
    ledOff();
}

void ledOn()
{
    digitalWrite(LED_BUILTIN, HIGH);
}

void ledOff()
{
    digitalWrite(LED_BUILTIN, LOW);
}