#include <WiFi.h>
#include <WiFiUdp.h>
#include <Adafruit_ST7735.h>

#define TFT_MOSI  13
#define TFT_SCLK  15
#define TFT_CS    27
#define TFT_DC    33
#define TFT_RST   32

Adafruit_ST7735 tft = Adafruit_ST7735(TFT_CS, TFT_DC, TFT_MOSI, TFT_SCLK, TFT_RST);

const char* ssid = "ESP_MONITOR_NET";
const char* password = "password123";
WiFiUDP udp;
unsigned int localPort = 1234;
char packetBuffer[512]; 

struct SystemStats {
  float cpuPct, cpuTemp, ramPct, ramGb;
  int cpuFreq;
  long diskRead, diskWrite, netDown;
  unsigned long uptime;
} stats;

unsigned long displayTimer = 0;
unsigned long lastPacketTime = 0;
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
  
  WiFi.softAP(ssid, password);
  udp.begin(localPort);

  drawWaitScreen();
}

void loop() {
  int packetSize = udp.parsePacket();
  if (packetSize) {
    int len = udp.read(packetBuffer, 511);
    if (len > 0) {
      packetBuffer[len] = 0;
      String input = String(packetBuffer);
      
      stats.cpuPct = getValue(input, "CP:").toFloat();
      stats.cpuTemp = getValue(input, "CT:").toFloat();
      stats.cpuFreq = getValue(input, "CF:").toInt();
      stats.ramPct = getValue(input, "RM:").toFloat();
      stats.ramGb = getValue(input, "RG:").toFloat();
      stats.diskRead = getValue(input, "DR:").toInt();
      stats.diskWrite = getValue(input, "DW:").toInt();
      stats.netDown = getValue(input, "ND:").toInt();
      stats.uptime = getValue(input, "UP:").toInt();
      
      lastPacketTime = millis();
    }
  }

  if (millis() - displayTimer > 5000) {
    displayTimer = millis();

    if (millis() - lastPacketTime > 7000) {
      stats.uptime = 0;
      drawWaitScreen();
    } else {
      drawUI();
    }
  }
}

void drawUI() {
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
      tft.setCursor(62, 35); tft.print("CPU");
      tft.setTextSize(1);
      tft.setCursor(35, 55); 
      tft.print("Load: "); tft.print(stats.cpuPct, 1); tft.print("%  ");
      tft.print(stats.cpuTemp, 0); tft.print("C");
      tft.drawRect(20, 70, 120, 12, ST77XX_WHITE);
      int fillW = map(constrain(stats.cpuPct, 0, 100), 0, 100, 0, 116);
      tft.fillRect(22, 72, fillW, 8, (stats.cpuPct > 80) ? ST77XX_RED : ST77XX_GREEN);
      tft.setCursor(62, 90);
      tft.setTextColor(ST77XX_YELLOW);
      tft.print(stats.cpuFreq); tft.print(" MHz");
      break;
    }
    case 1: { // --- RAM ---
      tft.setTextSize(2);
      tft.setTextColor(ST77XX_MAGENTA);
      tft.setCursor(62, 10); tft.print("RAM");
      tft.drawRect(65, 30, 30, 12, ST77XX_WHITE);
      for(int i=0; i<4; i++) tft.fillRect(68+(i*6), 38, 3, 3, ST77XX_WHITE);
      tft.setCursor(55, 50); tft.print(stats.ramPct, 1); tft.print("%");
      tft.drawRect(20, 75, 120, 12, ST77XX_WHITE);
      tft.fillRect(22, 77, map(constrain(stats.ramPct, 0, 100), 0, 100, 0, 116), 8, ST77XX_MAGENTA);
      tft.setTextSize(1);
      tft.setCursor(50, 95); tft.print("Used: "); tft.print(stats.ramGb, 1); tft.print(" GB");
      break;
    }
    case 2: { // --- NETWORK & DISK ---
      tft.setTextSize(2);
      tft.setTextColor(ST77XX_ORANGE);
      tft.setCursor(40, 10); tft.print("NETWORK");
      tft.setTextSize(1);
      tft.setTextColor(ST77XX_GREEN);
      tft.setCursor(35, 35); tft.print("DL: "); tft.print(stats.netDown); tft.print(" KB/s");
      tft.drawFastHLine(20, 55, 120, ST77XX_WHITE);
      tft.setTextSize(2);
      tft.setTextColor(ST77XX_WHITE);
      tft.setCursor(40, 65); tft.print("DISK IO");
      tft.setTextSize(1);
      tft.setCursor(30, 90); tft.print("Read: "); tft.print(stats.diskRead); tft.print(" KB/s");
      tft.setCursor(30, 105); tft.print("Write: "); tft.print(stats.diskWrite); tft.print(" KB/s");
      break;
    }
    case 3: { // --- UPTIME ---
      tft.fillScreen(0x000F);
      tft.setTextSize(2);
      tft.setTextColor(ST77XX_WHITE);
      tft.setCursor(45, 15); tft.print("UPTIME");
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

void drawWaitScreen() {
  tft.fillScreen(ST77XX_BLACK);
  tft.setTextSize(1);
  tft.drawRoundRect(17, 15, 139, 70, 8, ST77XX_CYAN);
  tft.setTextColor(ST77XX_CYAN);
  tft.setCursor(55, 25);
  tft.println("WiFi Ready");
  tft.setTextColor(ST77XX_WHITE);
  tft.setCursor(27, 40);
  tft.print("SSID: ");
  tft.println("ESP_MONITOR_NET");
  tft.setCursor(27, 52);
  tft.print("IP:   ");
  tft.println(WiFi.softAPIP());
  tft.setCursor(27, 64);
  tft.print("PSWD: ");
  tft.println("password123");
  tft.setTextColor(ST77XX_YELLOW);
  tft.setCursor(25, 100);
  tft.print("WAITING FOR DATA...");
}