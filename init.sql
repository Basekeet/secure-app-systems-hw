CREATE ROLE student LOGIN PASSWORD 'student_pass';
GRANT CONNECT ON DATABASE homework TO student;
GRANT USAGE ON SCHEMA public TO student;
