#include "M93Cx6.h"

M93Cx6::M93Cx6(uint8_t csPin, uint8_t skPin, uint8_t diPin, uint8_t doPin, uint8_t orgPin = 0xFF, uint8_t pinDelay = 0x64)
{
    M93Cx6(0xFF, csPin, skPin, diPin, doPin, orgPin, pinDelay);
}
M93Cx6::M93Cx6(uint8_t pwrPin, uint8_t csPin, uint8_t skPin, uint8_t diPin, uint8_t doPin, uint8_t orgPin = 0xFF, uint8_t pinDelay = 0x64)
{
    _pwrPin = pwrPin;
    _csPin = csPin;
    _skPin = skPin;
    _diPin = diPin;
    _doPin = doPin;
    _orgPin = orgPin;
    _pinDelay = pinDelay;
    _dataLength = 8;
    _org = ORG_8;
    _chip = M93C66;
    _addressLength = 9;
    _checkStatus = 1;

    if (_pwrPin != 0xFF)
        pinMode(_pwrPin, OUTPUT);
    powerDown();

    pinMode(_csPin, OUTPUT);
    pinMode(_skPin, OUTPUT);
    pinMode(_diPin, INPUT);
    pinMode(_doPin, OUTPUT);

    if (_orgPin != 0xFF)
        pinMode(_orgPin, OUTPUT);

    pinLow(_doPin);
    pinLow(_skPin);
    pinLow(_csPin);
    pinLow(_orgPin);
}

void M93Cx6::setChip(uint8_t chip)
{
    switch (chip)
    {
    case M93C46:
        _chip = chip;
        _addressLength = 6 + (_org == ORG_8 ? 1 : 0);
        break;
    case M93C56:
    case M93C66:
        _chip = chip;
        _addressLength = 8 + (_org == ORG_8 ? 1 : 0);
        break;
    case M93C76:
    case M93C86:
        _chip = chip;
        _addressLength = 10 + (_org == ORG_8 ? 1 : 0);
        break;
    }
}

void M93Cx6::setOrg(uint8_t org)
{
    if (_orgPin == 0xFF)
        return;
    switch (org)
    {
    case ORG_8:
        _org = org;
        _dataLength = 8;
        setChip(_chip);
        if (_orgPin != 0xFF)
            /*if (_orgPin != 0xFF) */ pinLow(_orgPin);
        break;
    case ORG_16:
        _org = org;
        _dataLength = 16;
        setChip(_chip);
        /*if (_orgPin != 0xFF) */ pinHigh(_orgPin);
        break;
    }
}

void M93Cx6::setPinDelay(uint8_t delay)
{
    _pinDelay = delay;
}

void M93Cx6::setCheckStatus(uint8_t status)
{
    _checkStatus = status;
}

void M93Cx6::powerUp()
{
    if (_pwrPin != 0xFF)
    {
        pinLow(_csPin);
        delayMicroseconds(_pinDelay);
        pinHigh(_pwrPin);
        delayMicroseconds(_pinDelay);
    }
}

void M93Cx6::powerDown()
{
    if (_pwrPin != 0xFF)
        pinLow(_pwrPin);
}

void M93Cx6::restart()
{
    powerDown();
    delay(10);
    powerUp();
}

uint16_t M93Cx6::read(uint16_t address)
{
    uint16_t data;
    pinHigh(_csPin);
    shiftOut(OP_READ, 3);
    shiftOut(address, _addressLength);
    data = shiftIn(_dataLength);
    pinLow(_csPin);
    delayMicroseconds(_pinDelay);
    return data;
}

void M93Cx6::write(uint16_t address, uint16_t data)
{
    pinHigh(_csPin);
    shiftOut(OP_WRITE, 3);
    shiftOut(address, _addressLength);
    shiftOut(data, _dataLength);
    pinLow(_csPin);
    delayMicroseconds(_pinDelay);
    checkStatus();
}

void M93Cx6::writeAll(uint8_t data)
{
    uint16_t out = (OP_WRAL << (_addressLength - 2));
    pinHigh(_csPin);
    shiftOut(out, _addressLength + 3);
    shiftOut(data, _dataLength);
    pinLow(_csPin);
    delayMicroseconds(_pinDelay);
    checkStatus();
}

void M93Cx6::erase(uint8_t address)
{
    pinHigh(_csPin);
    shiftOut(OP_ERASE, 3);
    shiftOut(address, _addressLength);
    delayMicroseconds(_pinDelay);
    checkStatus();
}

void M93Cx6::eraseAll()
{
    uint16_t out = (OP_ERAL << (_addressLength - 2));
    pinHigh(_csPin);
    shiftOut(out, _addressLength + 3);
    delayMicroseconds(_pinDelay);
    checkStatus();
}

void M93Cx6::writeEnable()
{
    uint16_t out = (OP_WEN << (_addressLength - 2));
    pinHigh(_csPin);
    shiftOut(out, _addressLength + 3);
    pinLow(_csPin);
    delayMicroseconds(_pinDelay);
}

void M93Cx6::writeDisable()
{
    uint16_t out = (OP_WDS << (_addressLength - 2));
    pinHigh(_csPin);
    shiftOut(out, _addressLength + 3);
    pinLow(_csPin);
    delayMicroseconds(_pinDelay);
}

uint16_t M93Cx6::shiftIn(uint16_t length)
{
    uint16_t value = 0;
    for (int i = 0; i < length; i++)
    {
        delayMicroseconds(_pinDelay);
        pinHigh(_skPin);
        delayMicroseconds(_pinDelay);
        value |= (digitalRead(_diPin) << ((length - 1) - i));
        pinLow(_skPin);
    }
    delayMicroseconds(_pinDelay);
    return value;
}

void M93Cx6::shiftOut(uint16_t data, uint8_t length)
{
    for (int i = 0; i < length; i++)
    {
        ((data & (1 << ((length - 1) - i))) ? pinHigh(_doPin) : pinLow(_doPin));
        delayMicroseconds(_pinDelay);
        pinHigh(_skPin);
        delayMicroseconds(_pinDelay);
        pinLow(_skPin);
    }
}

uint8_t M93Cx6::checkStatus()
{
    uint8_t r = 0;
    uint8_t quit = 0;
    uint32_t time = millis();

    if (_checkStatus)
    {
        pinHigh(_csPin);
        while (!quit)
        {
            r = digitalRead(_diPin);
            quit = (r || (millis() - time > CHECK_STATUS_TIMEOUT));
        }
        pinLow(_csPin);
    }
    return r;
}

void M93Cx6::pinHigh(uint8_t pin)
{
    if (pin < 8)
    {
        PORTD |= (B00000001 << pin);
    }
    else if (pin <= 13)
    {
        PORTB |= (B00000001 << (pin - 8));
    }
}

void M93Cx6::pinLow(uint8_t pin)
{
    if (pin < 8)
    {
        PORTD &= (~(B00000001 << pin));
    }
    else if (pin <= 13)
    {
        PORTB &= (~(B00000001 << (pin - 8)));
    }
}