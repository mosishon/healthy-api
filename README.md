# 🩺 Healthy API - مانیتورینگ پیشرفته سرویس‌ها

**Healthy API** یک ابزار قدرتمند و قابل توسعه برای مانیتورینگ لحظه‌ای سلامت (Health Check) وب‌سرویس‌های شماست. این پروژه با زبان **Go** نوشته شده و به شما کمک می‌کند تا با بررسی‌های دوره‌ای، از در دسترس بودن (Availability) و عملکرد صحیح سرویس‌هایتان مطمئن شوید و در صورت بروز هرگونه مشکل، بلافاصله از طریق کانال‌های مختلف (ایمیل و پیامک) با خبر شوید.

---

## ✨ امکانات کلیدی

- **مانیتورینگ چندین سرویس:** قابلیت تعریف و مانیتورینگ همزمان تعداد نامحدودی سرویس.
- **سیستم هشدار چند کاناله:** ارسال نوتیفیکیشن از طریق **ایمیل (SMTP)** و **پیامک (IPPanel)** با معماری قابل توسعه برای افزودن کانال‌های جدید.
- **بررسی‌های دوره‌ای هوشمند:** تنظیم بازه‌های زمانی دلخواه برای چک کردن هر سرویس.
- **جلوگیری از اسپم (Spam):** قابلیت تعریف یک دوره زمانی سکوت (`sleep_on_fail`) پس از شناسایی خطا برای جلوگیری از ارسال هشدارهای تکراری.
- **شرایط بررسی قابل تنظیم:** امکان تعریف **کد وضعیت HTTP** مورد انتظار (`expected_status_code`) برای هر سرویس.
- **اجرای همزمان (Concurrent):** استفاده از Goroutine برای مانیتورینگ تمام سرویس‌ها به صورت همزمان و بدون تداخل.
- **پیکربندی آسان:** تمام تنظیمات پروژه از طریق یک فایل `YAML` ساده و خوانا مدیریت می‌شود.

---

## 🚀 شروع به کار

### پیش‌نیازها

- **Go 1.21+**
- دسترسی به یک سرویس ایمیل (SMTP) یا پنل پیامک (مانند IPPanel)

### اجرا

۱. پروژه را Clone کنید:
```bash
git clone [https://github.com/mosishon/healthy-api.git](https://github.com/mosishon/healthy-api.git)
cd healthy-api
```

۲. یک فایل پیکربندی (مثلاً `config.yaml`) بر اساس نمونه زیر بسازید.

۳. برنامه را با دستور زیر اجرا کنید:
```bash
go run main.go -config=config.yaml
```

یا می‌توانید ابتدا فایل اجرایی را بسازید:
```bash
go build -o healthy-api
./healthy-api -config=config.yaml -verbose
```
> از فلگ `-verbose` برای دیدن لاگ‌های کامل برنامه استفاده کنید.

---

## ⚙️ پیکربندی (Configuration)

تمام تنظیمات در یک فایل YAML مدیریت می‌شوند. ساختار این فایل به شکل زیر است:

```yaml
services:
  - name: google-service-check # نام سرویس (برای نمایش در هشدارها)
    url: [https://google.com](https://google.com)
    targets:
      - notifier_id: personal_smtp # شناسه Notifier که در ادامه تعریف می‌شود
        recipients:
          - "your-email@example.com"
          - "another-email@example.com"
      - notifier_id: work_sms
        recipients:
          - "09123456789"
    check_period: 60        # هر ۶۰ ثانیه یک‌بار چک شود
    sleep_on_fail: 300      # بعد از شناسایی خطا، تا ۳۰۰ ثانیه بعدش دوباره چک نکن
    expected_status_code: 200 # کد وضعیت موفقیت‌آمیز

  - name: user-api
    url: [https://my-api.dev/health](https://my-api.dev/health)
    targets:
      - notifier_id: work_sms
        recipients:
          - "09120000000"
    check_period: 30
    sleep_on_fail: 120
    expected_status_code: 200

notifiers:
  ippanel: # لیست پنل‌های پیامک
    - id: work_sms
      url: <YOUR_IPPANEL_URL>
      user: <YOUR_IPPANEL_USERNAME>
      pass: <YOUR_IPPANEL_PASSWORD>

  smtp: # لیست سرورهای ایمیل
    - id: personal_smtp
      sender: "notifier@your-domain.com"
      password: "your-smtp-password"
      server: "smtp.your-domain.com"
      port: 587
```

---

## 🏗️ ساختار پروژه

معماری پروژه به صورت ماژولار طراحی شده تا به راحتی بتوان قابلیت‌های جدیدی به آن اضافه کرد.

```bash
.
├── config/         # منطق بارگذاری و پردازش فایل کانفیگ YAML
├── healthcheck/    # هسته اصلی برنامه برای اجرای حلقه‌های بررسی سرویس
├── model/          # تعریف ساختارها (Structs) مانند Service, Notifier, Config
├── notifier/       # سیستم ارسال هشدار (ایمیل، پیامک و...)
│   ├── notifier.go # اینترفیس اصلی برای Notifier ها
│   ├── registry.go # مدیریت و ثبت Notifier های مختلف
│   ├── mail.go     # پیاده‌سازی ارسال ایمیل (SMTP)
│   └── sms.go      # پیاده‌سازی ارسال پیامک (IPPanel)
├── main.go         # نقطه ورود و هماهنگ‌کننده ماژول‌ها
└── sample.yaml     # فایل نمونه پیکربندی
```

---

## 🗺️ نقشه راه آینده (Roadmap)

- [ ] افزودن **Graceful Shutdown** با استفاده از `context` برای مدیریت بهتر Goroutine ها.
- [ ] پیاده‌سازی **Unit Test** برای ماژول‌های `healthcheck` و `notifier`.
- [ ] پشتیبانی از **بررسی محتوای Response** با استفاده از عبارت‌های منظم (Regex).
- [ ] افزودن Notifier های بیشتر (مانند **Slack**, **Telegram**).
- [ ] ذخیره لاگ‌ها در یک فایل یا پایگاه داده برای تحلیل‌های بعدی.
- [ ] ساخت یک **رابط کاربری تحت وب (Web UI)** ساده برای نمایش وضعیت آنلاین سرویس‌ها.

---

## 🤝 مشارکت (Contributing)

از هرگونه مشارکت (PR و Issue) به شدت استقبال می‌شود! اگر ایده‌ای برای بهتر شدن پروژه دارید، خوشحال می‌شویم آن را با ما در میان بگذارید.

برای توسعه کد، لطفا اصول زیر را دنبال کنید:
- رعایت **قراردادهای نام‌گذاری (Naming Conventions)** در Go.
- طراحی مبتنی بر **اینترفیس (Interface-based Design)** برای انعطاف‌پذیری بیشتر.
- استفاده از **لاگر (Logger)** قابل کنترل برای دیباگ بهتر.

---

## 📄 لایسنس

**Mostafa Arshadi** (با افتخار، برای یادگیری، پیشرفت و کار تیمی ❤️)