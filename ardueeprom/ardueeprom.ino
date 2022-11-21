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

M93Cx6 ep = M93Cx6(PWR_PIN, CS_PIN, SK_PIN, DO_PIN, DI_PIN, ORG_PIN, 100);

int chipAddr = 0x08;
int orgAddr = 0x10;
int sizeAddr = 0x20;
int delayAddr = 0x30;

static uint8_t cfgChip;
static uint16_t cfgSize;
static uint8_t cfgOrg;
static uint8_t cfgDelay;

void setup()
{
    Serial.begin(57600);

#ifdef USE_EEPROM
    // load settings from eeprom
    loadSettings();
#endif

    pinMode(LED_BUILTIN, OUTPUT); // LED
    while (!Serial)
    {
        delay(20); // wait for serial port to connect. Needed for native USB
    }
    Serial.println();
}

void loop()
{
    handleSerial();
}

#ifdef USE_EEPROM
void loadSettings()
{
    if (EEPROM.read(0x00) == 0x20)
    {
        cfgChip = EEPROM.read(chipAddr);
        EEPROM.get(sizeAddr, cfgSize);
        cfgOrg = EEPROM.read(orgAddr);
        ep.setChip(cfgChip);
        ep.setOrg(cfgOrg);
        ep.setPinDelay(cfgDelay);
    }
}
#endif

static uint8_t bufferLength;       // number of characters currently in the buffer
const uint8_t BUFF_SIZE = 16;      // make it big enough to hold your longest command
static char buffer[BUFF_SIZE + 1]; // +1 allows space for the null terminator

void handleSerial()
{
    if (Serial.available())
    {
        char c = Serial.read();
        if ((c == '\r' && bufferLength == 0))
        {
            help();
            return;
        }

        if ((c == '\r') || (c == '\n'))
        {
            // end-of-line received
            if (bufferLength > 0)
            {
                handleCmd(buffer);
            }
            bufferLength = 0;
            return;
        }

        if (c == 127) // handle backspace during command input
        {
            if (bufferLength > 0)
            {
                bufferLength--;
            }
            return;
        }

        if (bufferLength < BUFF_SIZE)
        {
            buffer[bufferLength++] = c;  // append the received character to the array
            buffer[bufferLength] = 0x00; // append the null terminator
        }
        else
        {
            Serial.write('\a');
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
    case 'x':
        break;
    case '?':
        settings();
        return;
    default:
        Serial.println("invalid command");
        help();
        return;
    }

    if (bufferLength > 8) // parse long command with set options inline
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

    pos = strtok(NULL, ",");
    cfgDelay = atoi(pos);

    if (cfgSize == 0)
    {
        Serial.println("\ainvalid size");
        return;
    }

    if (!setChip())
    {
        return;
    }

    if (!setOrg())
    {
        return;
    }

    if (!setDelay())
    {
        return;
    }

#ifdef USE_EEPROM
    EEPROM.put(chipAddr, cfgChip);
    EEPROM.put(orgAddr, cfgOrg);
    EEPROM.put(sizeAddr, cfgSize);
    EEPROM.put(delayAddr, cfgDelay);
    EEPROM.put(0x00, 0x20); // if this is not 0x20 settings will not be loaded from eeprom
#endif
}

bool setChip()
{
    switch (cfgChip)
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
        Serial.println("\ainvalid chip");
        return false;
    }

    return true;
}

bool setOrg()
{
    switch (cfgOrg)
    {
    case 8:
        ep.setOrg(ORG_8);
        break;
    case 16:
        ep.setOrg(ORG_16);
        break;
    default:
        Serial.println("\ainvalid org");
        return false;
    }

    return true;
}

bool setDelay()
{
    ep.setPinDelay(cfgDelay);
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
    Serial.print("delay: ");
    Serial.println(cfgDelay);
}

void read()
{
    delay(10);
    uint16_t c = 0;
    for (uint16_t i = 0; i < cfgSize; i++)
    {
        c = ep.read(i);
        switch (cfgOrg)
        {
        case 8:
            Serial.write(c);
            break;
        case 16:
            Serial.write(c >> 8);
            Serial.write(c & 0xFF);
            break;
        }
    }
    Serial.println();
}

static unsigned long lastData;

void write()
{
    lastData = millis();
    ep.writeEnable();
    uint8_t buff[2];
    uint8_t pos = 0;

    Serial.write('\f');
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
        switch (cfgOrg)
        {
        case 8:
            ep.write(i, Serial.read());
            break;

        case 16:
            if (pos == 2)
            {
                uint16_t wd = ((uint16_t)buff[1] << 8) | buff[0];
                ep.write(i, wd);
                pos = 0;
                break;
            }
            buff[pos++] = Serial.read();
            i--;
            break;
        }

        lastData = millis();
        Serial.print("\f");
    }
    ep.writeDisable();
    Serial.println("\r\n--- write done ---");
}

void erase()
{
    ep.writeEnable();
    ep.eraseAll();
    ep.writeDisable();
    Serial.println("\aeeprom erased");
}

static uint8_t linePos;
void printBin()
{
    linePos = 0;
    char buf[4];
    ledOn();
    Serial.println("--- Hex dump ---");
    uint16_t c = 0;
    for (uint16_t i = 0; i < cfgSize; i++)
    {
        switch (cfgOrg)
        {
        case 8:
            c = ep.read(i);
            sprintf(buf, "%02X ", c);
            break;
        case 16:
            c = ep.read(i);
            sprintf(buf, "%02X%02X ", c >> 8, c & 0xFF);
            break;
        }
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
