#include <Arduino.h>
#include "M93Cx6.h"
#include <CmdBuffer.hpp>

#define PWR_PIN 9
#define CS_PIN 10
#define SK_PIN 7
#define DI_PIN 11
#define DO_PIN 12
#define ORG_PIN 8

M93Cx6 ep = M93Cx6(PWR_PIN, CS_PIN, SK_PIN, DO_PIN, DI_PIN, ORG_PIN);

int incomingByte = 0;
bool debug = false;

void setup()
{
    Serial.begin(57600);
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

void handleSerial()
{
    const int BUFF_SIZE = 32;          // make it big enough to hold your longest command
    static char buffer[BUFF_SIZE + 1]; // +1 allows space for the null terminator
    static int length = 0;             // number of characters currently in the buffer

    if (Serial.available())
    {
        char c = Serial.read();
        if ((c == '\r') || (c == '\n'))
        {
            // end-of-line received
            if (length > 0)
            {
                handleReceivedMessage(buffer);
            }
            length = 0;
        }
        else
        {
            if (length < BUFF_SIZE)
            {
                buffer[length++] = c; // append the received character to the array
                buffer[length] = 0;   // append the null terminator
            }
            else
            {
                Serial.write('\a');
            }
        }
    }
}

char *cmd;
uint8_t chip;
uint16_t size;
uint8_t org;

void handleReceivedMessage(char *msg)
{
    switch (msg[0])
    {
    case 'h':
        help();
        return;
    case 'w':
    case 'r':
    case 'e':
    case 'p':
        break;

    default:
        Serial.println("invalid command");
        help();
        return;
    }

    chip = 0;
    size = 0;
    org = 0;

    parseMsg(msg);

    if (!setChip(chip))
    {
        return;
    }

    if (!setOrg(org))
    {
        return;
    }

    ep.powerUp();
    delayMicroseconds(100);

    switch (msg[0])
    {
    case 'r':
        read(size);
        break;
    case 'w':
        write(size);
        break;
    case 'p':
        printBin(size);
        break;
    case 'e':
        erase();
        break;
    }
}

void parseMsg(char *msg)
{
    char *pos;
    cmd = strtok(msg, ",");

    pos = strtok(NULL, ",");
    chip = atoi(pos);

    pos = strtok(NULL, ",");
    size = atoi(pos);

    pos = strtok(NULL, ",");
    org = atoi(pos);

    if (size == 0)
    {
        Serial.println("invalid size");
        return;
    }
    if (chip == 0)
    {
        Serial.println("invalid chip");
        return;
    }
    if (org == 0)
    {
        Serial.println("invalid org");
        return;
    }
    if (debug)
    {
        Serial.println("");
        Serial.print("cmd: ");
        Serial.print(cmd);
        Serial.print(", chip: ");
        Serial.print(chip);
        Serial.print(", size: ");
        Serial.print(size);
        Serial.print(", org ");
        Serial.println(org);
    }
}

bool setChip(uint8_t chip)
{
    switch (chip)
    {
    case 44:
        ep.setChip(M93C46);
        return true;
    case 55:
        ep.setChip(M93C56);
        return true;
    case 66:
        ep.setChip(M93C66);
        return true;
    case 76:
        ep.setChip(M93C76);
        return true;
    case 86:
        ep.setChip(M93C86);
        return true;
    default:
        Serial.println("invalid CHIP");
        return false;
    }
}

bool setOrg(uint8_t org)
{
    switch (org)
    {
    case 8:
        ep.setOrg(ORG_8);
        return true;
    case 16:
        ep.setOrg(ORG_16);
        return true;
    default:
        Serial.println("invalid ORG");
        return false;
    }
}

void help()
{
    Serial.println("--- Arep ---");
    Serial.println("r,<chip>,<size>,<org> - Read eeprom");
    Serial.println("w,<chip>,<size>,<org> - Initiate write mode");
    Serial.println("e,<chip>,<size>,<org> - Erase eeprim");
    Serial.println("p,<chip>,<size>,<org> - Hex print eeprom content");
    Serial.println("h - This help");
}

void read(size_t size)
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

void write(size_t size)
{
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

        uint8_t c = Serial.read();
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

void printBin(size_t size)
{
    uint8_t pos = 0;
    uint8_t c;
    char buf[3];
    ledOn();
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