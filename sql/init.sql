CREATE TABLE IF NOT EXISTS employees (
 id SERIAL PRIMARY KEY, name VARCHAR(100) NOT NULL, department VARCHAR(100) NOT NULL,
 position VARCHAR(100) NOT NULL, salary NUMERIC(12,2) NOT NULL CHECK(salary >= 0),
 is_active BOOLEAN NOT NULL DEFAULT TRUE
);
INSERT INTO employees(name,department,position,salary,is_active)
SELECT s.name,s.department,s.position,s.salary,s.is_active FROM (VALUES
 ('Иван Петров','IT','Developer',450000,true), ('Анна Сидорова','HR','HR Manager',350000,true),
 ('Алексей Иванов','IT','Team Lead',650000,true), ('Мария Смирнова','Finance','Accountant',400000,true),
 ('Дмитрий Ким','IT','QA Engineer',380000,true)
) AS s(name,department,position,salary,is_active) WHERE NOT EXISTS (SELECT 1 FROM employees);
