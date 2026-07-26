-- Pengembalian batasan validasi ke kondisi semula.
-- Catatan: Apabila terdapat entri dengan question_type='file_upload',
-- perintah pembaruan di bawah tidak akan mengubahnya karena tidak terdapat
-- representasi yang setara pada skema sebelumnya.
UPDATE dynamic_questions SET question_type = 'select' WHERE question_type = 'dropdown';

ALTER TABLE dynamic_questions DROP CONSTRAINT IF EXISTS dynamic_questions_question_type_check;

ALTER TABLE dynamic_questions ADD CONSTRAINT dynamic_questions_question_type_check
    CHECK (question_type IN ('text', 'textarea', 'number', 'date', 'select', 'radio', 'checkbox'));
