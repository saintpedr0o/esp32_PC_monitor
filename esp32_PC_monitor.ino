#include "BluetoothSerial.h"
#include <Adafruit_GFX.h>
#include <Adafruit_ST7735.h>
#include <SPI.h>

#define TFT_MOSI  13 // sda
#define TFT_SCLK  15
#define TFT_CS    4
#define TFT_DC    2
#define TFT_RST   5

// Adafruit_ST7735 tft = Adafruit_ST7735(CS_PIN, DC_PIN, RST_PIN);
Adafruit_ST7735 tft = Adafruit_ST7735(TFT_CS, TFT_DC, TFT_MOSI, TFT_SCLK, TFT_RST);

BluetoothSerial SerialBT;

struct SystemStats {
  float cpuPct, cpuTemp, ramPct, ramGb;
  int cpuFreq;
  long diskRead, diskWrite, netDown;
  unsigned long uptime;
} stats;

String inputBuffer = "";
unsigned long displayTimer = 0;
int page = 0;

String getValue(String data, String key) {
  int keyPos = data.indexOf(key);
  if (keyPos == -1) return "0";
  int start = keyPos + key.length();
  int end = data.indexOf("|", start);
  if (end == -1) end = data.indexOf("\n", start);
  return data.substring(start, end);
}

void setup() {
  tft.initR(INITR_BLACKTAB); 
  tft.setRotation(1); 
  tft.fillScreen(ST77XX_BLACK);
  
  tft.setCursor(60, 60);
  tft.setTextColor(ST77XX_WHITE);
  tft.setTextSize(1);
  tft.println("Booting...");

  // try start Bluetooth
  if(!SerialBT.begin("ESP_MONITOR")){
    tft.fillScreen(ST77XX_RED);
    tft.setCursor(10, 50);
    tft.setTextColor(ST77XX_BLACK);
    tft.setTextSize(2);
    tft.println("BT ERROR!");
    while(1);
  }

  drawBluetoothWaitScreen();

  esp_bt_cod_t cod;
  cod.major = 0b00101;
  cod.minor = 0b000000;
  cod.service = 0b00000100000;
  esp_bt_gap_set_cod(cod, ESP_BT_SET_COD_ALL);
}

void loop() {
  // read Bluetooth data
  while (SerialBT.available()) {
    char c = SerialBT.read();
    if (c == '\n') {
      // handshake
      if (inputBuffer == "IDENTIFY") {
        SerialBT.println("ESP_MONITOR_READY"); 
      } 
      else if (inputBuffer.length() > 0) {
        stats.cpuPct = getValue(inputBuffer, "CP:").toFloat();
        stats.cpuTemp = getValue(inputBuffer, "CT:").toFloat();
        stats.cpuFreq = getValue(inputBuffer, "CF:").toInt();
        stats.ramPct = getValue(inputBuffer, "RM:").toFloat();
        stats.ramGb = getValue(inputBuffer, "RG:").toFloat();
        stats.diskRead = getValue(inputBuffer, "DR:").toInt();
        stats.diskWrite = getValue(inputBuffer, "DW:").toInt();
        stats.netDown = getValue(inputBuffer, "ND:").toInt();
        stats.uptime = getValue(inputBuffer, "UP:").toInt();
      }
      inputBuffer = "";
    } else {
      inputBuffer += c;
    }
}

  // main view logic
  if (millis() - displayTimer > 5000) {
    displayTimer = millis();

    // Bluetooth conneciton check
    if (!SerialBT.hasClient()) {
      drawBluetoothWaitScreen();
      stats.uptime = 0; // Resetting stats for a fresh start on reconnect
    } 
    // If connected but no data (uptime still 0)
    else if (stats.uptime == 0) {
      tft.fillScreen(ST77XX_BLACK);
      tft.setCursor(40, 60);
      tft.setTextColor(ST77XX_YELLOW);
      tft.setTextSize(1);
      tft.print("CONNECTED");
      tft.setCursor(30, 70);
      tft.print("WAITING DATA...");
    } 
    // If data coming - show monitoring pages
    else {
      tft.fillScreen(ST77XX_BLACK); 
      switch (page) {
        case 0: { // --- CPU ---
          int iconX = 70; 
          tft.drawRect(iconX, 8, 20, 20, ST77XX_WHITE);
          for(int i=0; i<20; i+=5) {
            tft.drawLine(iconX-5, 10+i, iconX, 10+i, ST77XX_WHITE); 
            tft.drawLine(iconX+20, 10+i, iconX+25, 10+i, ST77XX_WHITE); 
          }
          tft.setTextSize(2);
          tft.setTextColor(ST77XX_CYAN);
          tft.setCursor(62, 35);
          tft.print("CPU");
          tft.setTextSize(1);
          tft.setCursor(35, 55); 
          tft.print("Load: "); tft.print(stats.cpuPct, 1); tft.print("%  ");
          tft.print(stats.cpuTemp, 0); tft.print("C");
          int barW = 120;
          int barX = 20;
          tft.drawRect(barX, 70, barW, 12, ST77XX_WHITE);
          int fillW = map(stats.cpuPct, 0, 100, 0, barW - 4);
          uint16_t cpuCol = (stats.cpuPct > 80) ? ST77XX_RED : ST77XX_GREEN;
          tft.fillRect(barX + 2, 72, fillW, 8, cpuCol);
          tft.setCursor(62, 90);
          tft.setTextColor(ST77XX_YELLOW);
          tft.print(stats.cpuFreq); tft.print(" MHz");
          break;
        }
        case 1: { // --- RAM ---
          tft.setTextSize(2);
          tft.setTextColor(ST77XX_MAGENTA);
          tft.setCursor(62, 10);
          tft.print("RAM");
          tft.drawRect(65, 30, 30, 12, ST77XX_WHITE);
          for(int i=0; i<4; i++) tft.fillRect(68+(i*6), 38, 3, 3, ST77XX_WHITE);
          tft.setTextSize(2);
          tft.setCursor(55, 50);
          tft.print(stats.ramPct, 1); tft.print("%");
          int barW = 120;
          int barX = 20;
          tft.drawRect(barX, 75, barW, 12, ST77XX_WHITE);
          int fillW = map(stats.ramPct, 0, 100, 0, barW - 4);
          tft.fillRect(barX + 2, 77, fillW, 8, ST77XX_MAGENTA);
          tft.setTextSize(1);
          tft.setCursor(50, 95);
          tft.print("Used: "); tft.print(stats.ramGb, 1); tft.print(" GB");
          break;
        }
        case 2: { // --- NETWORK ---
          tft.setTextSize(2);
          tft.setTextColor(ST77XX_ORANGE);
          tft.setCursor(40, 10);
          tft.print("NETWORK");
          tft.setTextSize(1);
          tft.setTextColor(ST77XX_GREEN);
          tft.setCursor(35, 35);
          tft.print("DL: "); tft.print(stats.netDown); tft.print(" KB/s");
          tft.drawFastHLine(20, 55, 120, ST77XX_WHITE);
          tft.setTextSize(2);
          tft.setTextColor(ST77XX_WHITE);
          tft.setCursor(40, 65);
          tft.print("DISK IO");
          tft.setTextSize(1);
          tft.setCursor(30, 90);
          tft.print("Read: "); tft.print(stats.diskRead); tft.print(" KB/s");
          tft.setCursor(30, 105);
          tft.print("Write: "); tft.print(stats.diskWrite); tft.print(" KB/s");
          break;
        }
        case 3: { // --- UPTIME ---
          tft.fillScreen(0x000F);
          tft.setTextSize(2);
          tft.setTextColor(ST77XX_WHITE);
          tft.setCursor(45, 15);
          tft.print("UPTIME");
          tft.drawCircle(80, 65, 20, ST77XX_WHITE);
          tft.drawLine(80, 65, 80, 52, ST77XX_WHITE); 
          tft.drawLine(80, 65, 92, 65, ST77XX_WHITE); 
          tft.setTextSize(1);
          tft.setCursor(50, 100);
          tft.print(stats.uptime / 3600); tft.print("h ");
          tft.print((stats.uptime % 3600) / 60); tft.print("m ");
          tft.print(stats.uptime % 60); tft.print("s");
          break;
        }
      }
      page = (page + 1) % 4;
    }
  }
}

void drawBluetoothWaitScreen() {
  tft.fillScreen(ST77XX_BLACK);
  
  int bx = 80; 
  int by = 45; 
  int s = 15;

  tft.drawRoundRect(bx - 20, by - 22, 40, 44, 15, ST77XX_CYAN);
  
  // Bluetooth icon
  tft.drawLine(bx, by - s, bx, by + s, ST77XX_CYAN); 
  tft.drawLine(bx, by - s, bx + s/2, by - s/2, ST77XX_CYAN); 
  tft.drawLine(bx + s/2, by - s/2, bx - s/2, by + s/2, ST77XX_CYAN); 
  tft.drawLine(bx - s/2, by - s/2, bx + s/2, by + s/2, ST77XX_CYAN); 
  tft.drawLine(bx + s/2, by + s/2, bx, by + s, ST77XX_CYAN);
 
  tft.setTextSize(1);
  tft.setTextColor(ST77XX_WHITE);
  tft.setCursor(30, 85);
  tft.print("WAITING CONNECTION");
  
  tft.setCursor(40, 100);
  tft.setTextColor(ST77XX_CYAN);
  tft.print("ID: ESP_MONITOR");
}
