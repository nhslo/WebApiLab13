# Лабораторная работа 13 — сотрудники и PostgreSQL

## Краткий отчёт

Учет сотрудников в PostgreSQL с фильтром по подразделению и статистикой.

## Выполнение по заданию

### 1. База данных

Таблица `employees` и пять сотрудников создаются скриптом `sql/init.sql`. Доступ к БД выполняется через `database/sql` и pgx.

### 2. Маршруты

GET и POST `/employees`, GET, PUT и DELETE `/employees/{id}`, фильтр `?department=IT`, GET `/employees/statistics`.

### 3. Проверка

Проверены добавление, изменение, удаление, фильтр IT, статистика и отрицательная зарплата.

## Запуск

Локальный PostgreSQL установлен в `%LOCALAPPDATA%\Programs\postgres-labs`, порт 55432. База `employee_db` должна быть создана; выполните скрипт sql/init.sql. Откройте PowerShell в этой папке:

```powershell
$env:DATABASE_URL = 'postgres://postgres:YOUR_POSTGRES_PASSWORD@127.0.0.1:55432/employee_db?sslmode=disable'
go run .
```

Замените `YOUR_POSTGRES_PASSWORD` паролем своей базы (специальные символы в URL нужно закодировать). Откройте нужный адрес на `http://localhost:8080`. Если PostgreSQL остановлен, запустите его:

```powershell
& "$env:LOCALAPPDATA\Programs\postgres-labs\pgsql\bin\pg_ctl.exe" -D "$env:LOCALAPPDATA\Programs\postgres-labs\data" -l "$env:LOCALAPPDATA\Programs\postgres-labs\server.log" -o "-h 127.0.0.1 -p 55432" -w start
```

На другом компьютере создайте базу `employee_db`, выполните `sql/init.sql` через `psql` и задайте свою строку подключения в `DATABASE_URL`.

## Пример добавления

Пока сервер работает, откройте второе окно PowerShell:

```powershell
$body = @{ name="Сергей Орлов"; department="IT"; position="Developer"; salary=500000; isActive=$true } | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri http://localhost:8080/employees -ContentType "application/json; charset=utf-8" -Body ([System.Text.Encoding]::UTF8.GetBytes($body))
```

## Скриншот выполнения

![Результат работы приложения](docs/result.png)

## Вывод

SQL-запросы используют параметры. Статистика считается в PostgreSQL с помощью COUNT и AVG.
