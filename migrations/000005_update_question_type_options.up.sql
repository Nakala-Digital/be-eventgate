-- Pembaruan batasan validasi untuk question_type pada tabel dynamic_questions.
-- Batasan sebelumnya hanya mengizinkan: 'text', 'textarea', 'number', 'date', 'select', 'radio', 'checkbox'.
-- Pembaruan ini mengubah 'select' menjadi 'dropdown' dan menambahkan opsi 'file_upload'.
--
-- Catatan: Nama batasan (dynamic_questions_question_type_check) merupakan penamaan bawaan
-- yang dihasilkan oleh PostgreSQL. Harap sesuaikan perintah DROP CONSTRAINT di bawah
-- apabila nama batasan pada basis data berbeda.

-- 1. Migrasi data lama apabila terdapat entri dengan question_type='select'
UPDATE dynamic_questions SET question_type = 'dropdown' WHERE question_type = 'select';

-- 2. Penghapusan batasan lama
ALTER TABLE dynamic_questions DROP CONSTRAINT IF EXISTS dynamic_questions_question_type_check;

-- 3. Penambahan batasan baru dengan opsi dropdown dan file_upload
ALTER TABLE dynamic_questions ADD CONSTRAINT dynamic_questions_question_type_check
    CHECK (question_type IN ('text', 'textarea', 'number', 'date', 'dropdown', 'radio', 'checkbox', 'file_upload'));
