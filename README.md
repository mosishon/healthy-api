# 🩺 Healthy API - Service Health Checker in Go

پروژه‌ای برای بررسی در دسترس بودن (Availability) وب‌سرویس‌ها به صورت دوره‌ای و ارسال هشدار از طریق پیامک در صورت بروز مشکل.

---

## 🚀 امکانات اصلی

- بررسی وضعیت سرویس‌های مشخص‌شده (HTTP Status Code)
- فقط از طریق HTTPS
- اجرای دوره‌ای (Periodic)
- پشتیبانی از Sleep بین خطاها برای جلوگیری از spam شدن
- ارسال هشدار به شماره‌های مشخص‌شده از طریق پیامک
- پشتیبانی از چندین سرویس هم‌زمان (Multi-Service)
- ساختار قابل گسترش، انعطاف‌پذیر و تست‌پذیر

---

## 📦 ساختار پروژه

```bash
.
├── config/         # منطق بارگذاری کانفیگ YAML
├── healthcheck/    # اجرای حلقه‌های چک کردن سرویس
├── model/          # ساختارها (structs) شامل Service، Notification، Config
├── notifier/       # سیستم ارسال پیامک
├── main.go         # نقطه ورود CLI برنامه
└── sample.yaml     # فایل نمونه پیکربندی
```

---

## 🛠️ پیش‌نیازها

- Go 1.21+
- دسترسی به API پیامک (مثل ippanel)

---

## ⚙️ اجرای برنامه

```bash
go run main.go -config=sample.yaml
```

یا بعد از build:

```bash
go build -o healthy-api
./healthy-api -config=sample.yaml
```

---

## 🧾 قالب فایل پیکربندی (YAML)

```yaml
user: your_username_here
pass: your_password_here
services:
  - name: auth-service
    url: https://auth.example.com/health
    phones:
      - "+989123456789"
      - "+989987654321"
    check_period: 30        # هر ۳۰ ثانیه یکبار چک کن
    sleep_on_fail: 180      # اگر داون بود، تا ۳ دقیقه دیگه چک نکن
    expected_status_code: 200
  - name: payment-service
    url: https://pay.example.com/ping
    phones:
      - "+989123456789"
    check_period: 60
    sleep_on_fail: 300
    expected_status_code: 200
```

---

## 📤 ساختار نوتیفیکیشن پیامک

در صورت داون بودن یک سرویس، پیامی با الگوی مشخص به شماره‌های تنظیم‌شده ارسال می‌شود.  
پشتیبانی از چند سرویس پیامکی در آینده قابل افزودن است.

---

## ✅ TODO / نقشه راه آینده

- [ ] افزودن ایمیل به‌جای یا همراه با پیامک
- [ ] اضافه کردن context برای graceful shutdown
- [ ] تست یونیت برای ماژول‌های healthcheck و notifier
- [ ] اضافه کردن web UI برای مشاهده وضعیت سرویس‌ها
- [ ] استفاده از یک پایگاه داده یا فایل برای ذخیره لاگ‌ها

---

## 🤝 مشارکت

PR و Issue خوشحال‌مون می‌کنه :)  
برای توسعه کد سعی کن از اصول زیر پیروی کنی:
- رعایت naming convention های Go
- استفاده از context و logger های قابل کنترل
- طراحی بر اساس interface

---

## 📄 لایسنس

**mosTafa Arshadi**  
(با افتخار توسعه داده شده برای یادگیری، بهره‌وری و پیشرفت تیمی ❤️)