-- GENERATED FILE. Source: front/mock-api/data.js.
-- Regenerate with: node scripts/generate-mock-seed.mjs
-- All mock accounts use password: password123.

BEGIN;

INSERT INTO categories (category_id, name, parent_id, created_at, updated_at) VALUES
('7eab0b28-c44c-5911-8522-176850b28be6', 'Товары для компьютера', NULL, '2026-08-07T00:00:00Z', '2026-08-07T00:00:00Z'),
('1821377d-fbf3-55c6-bd88-2a36070eac2d', 'Комплектующие', '7eab0b28-c44c-5911-8522-176850b28be6', '2026-08-07T00:00:00Z', '2026-08-07T00:00:00Z'),
('4a82b1a2-c73c-505e-91ba-e1cbc3b4ad25', 'Видеокарты', '1821377d-fbf3-55c6-bd88-2a36070eac2d', '2026-08-07T00:00:00Z', '2026-08-07T00:00:00Z'),
('85ca78e5-1b8f-5de2-a0af-41d06874d1d1', 'Игры для приставок', NULL, '2026-08-07T00:00:00Z', '2026-08-07T00:00:00Z')
ON CONFLICT DO NOTHING;


INSERT INTO customers (customer_id, email, password_hash, created_at, updated_at, is_active) VALUES
('4f7a8183-d03c-52ad-9ef9-9821a1f40c8b', 'alexey@example.com', '$2y$10$7fpSvEXlg0/63LutgOzEyeCMYUyegqEZ3P67sZsN0y0khT0JCghMy', '2026-08-07T00:00:00Z', '2026-08-07T00:00:00Z', TRUE),
('549fe311-ecdd-5f4e-9c1d-cea2d100e286', 'maria@example.com', '$2y$10$7fpSvEXlg0/63LutgOzEyeCMYUyegqEZ3P67sZsN0y0khT0JCghMy', '2026-08-07T00:00:00Z', '2026-08-07T00:00:00Z', TRUE),
('42bcc017-c6ef-5eb9-898f-6ff0d01293b2', 'ivan@example.com', '$2y$10$7fpSvEXlg0/63LutgOzEyeCMYUyegqEZ3P67sZsN0y0khT0JCghMy', '2026-08-07T00:00:00Z', '2026-08-07T00:00:00Z', TRUE),
('362a3db9-96a2-5fb4-925b-c8500cf395c1', 'olga@example.com', '$2y$10$7fpSvEXlg0/63LutgOzEyeCMYUyegqEZ3P67sZsN0y0khT0JCghMy', '2026-08-07T00:00:00Z', '2026-08-07T00:00:00Z', TRUE),
('5e96d7bb-c76c-5558-881e-1b132e49d342', 'dmitry@example.com', '$2y$10$7fpSvEXlg0/63LutgOzEyeCMYUyegqEZ3P67sZsN0y0khT0JCghMy', '2026-08-07T00:00:00Z', '2026-08-07T00:00:00Z', TRUE),
('f672cf38-b039-595d-84b3-821e9d5b3af7', 'elena@example.com', '$2y$10$7fpSvEXlg0/63LutgOzEyeCMYUyegqEZ3P67sZsN0y0khT0JCghMy', '2026-08-07T00:00:00Z', '2026-08-07T00:00:00Z', TRUE),
('d3b90730-bf1f-5c12-95c7-b1ff3908167c', 'sergey@example.com', '$2y$10$7fpSvEXlg0/63LutgOzEyeCMYUyegqEZ3P67sZsN0y0khT0JCghMy', '2026-08-07T00:00:00Z', '2026-08-07T00:00:00Z', TRUE),
('f69088b0-d6a9-5f50-bb4e-be9b46cb8664', 'natalia@example.com', '$2y$10$7fpSvEXlg0/63LutgOzEyeCMYUyegqEZ3P67sZsN0y0khT0JCghMy', '2026-08-07T00:00:00Z', '2026-08-07T00:00:00Z', TRUE),
('1a9b30df-8e74-53f8-a55d-0c8a016995be', 'pavel@example.com', '$2y$10$7fpSvEXlg0/63LutgOzEyeCMYUyegqEZ3P67sZsN0y0khT0JCghMy', '2026-08-07T00:00:00Z', '2026-08-07T00:00:00Z', TRUE),
('b6ead879-5ff3-5f09-af92-0003c53e4465', 'irina@example.com', '$2y$10$7fpSvEXlg0/63Lut9a1f40c8b', '2026-08-07T00:00:00Z', '2026-08-07T00:00:00Z', TRUE),
('2db05252-81a6-5e50-b52f-57a19da8baa7', 'roman@example.com', '$2y$10$7fpSvEXlg0/63LutgOzEyeCMYUyegqEZ3P67sZsN0y0khT0JCghMy', '2026-08-07T00:00:00Z', '2026-08-07T00:00:00Z', TRUE)
ON CONFLICT DO NOTHING;


INSERT INTO products (
    product_id,
    customer_id,
    category_id,
    title,
    description,
    image,
    price,
    location,
    status,
    created_at,
    updated_at
) VALUES
(
    '1bb647e1-a136-5c68-9ad8-f7c3b880816b',
    '4f7a8183-d03c-52ad-9ef9-9821a1f40c8b',
    '4a82b1a2-c73c-505e-91ba-e1cbc3b4ad25',
    'Видеокарта GeForce RTX 3060 12 ГБ',
    'В отличном состоянии, использовалась для игр. Полный комплект, проверка при встрече.',
    'https://50.img.avito.st/image/1/1.iedNrLa4JQ57G6cDU56eolkMJwjzDacYewAnDP0FLQT7._xzbVJNce46KNeP-4N3tbbh2TTVb5eNgmHDUYIyRsnU',
    28990,
    'Псков',
    'active',
    '2026-08-06T18:40:00Z',
    '2026-08-06T18:40:00Z'
),
(
    '27d2d5c6-d819-51a8-a730-629f37d05784',
    '549fe311-ecdd-5f4e-9c1d-cea2d100e286',
    '4a82b1a2-c73c-505e-91ba-e1cbc3b4ad25',
    'AMD Radeon RX 6600 8GB',
    'Тихая игровая видеокарта, не майнилась. Возможен обмен на консоль или игры.',
    'https://images.unsplash.com/photo-1591488320449-011701bb6704?auto=format&fit=crop&w=900&q=80',
    21500,
    'Псков',
    'active',
    '2026-08-05T12:15:00Z',
    '2026-08-05T12:15:00Z'
),
(
    '71aa4523-dca9-5f8e-9cb0-b448765d8c84',
    '42bcc017-c6ef-5eb9-898f-6ff0d01293b2',
    '85ca78e5-1b8f-5de2-a0af-41d06874d1d1',
    'Marvel''s Spider-Man 2 для PS5',
    'Физический диск, оригинальное издание. Диск и коробка без повреждений.',
    'https://images.unsplash.com/photo-1605899435973-ca2d1a8861cf?auto=format&fit=crop&w=900&q=80',
    3990,
    'Псков',
    'active',
    '2026-08-07T09:30:00Z',
    '2026-08-07T09:30:00Z'
),
(
    'a2adc3ce-b425-5671-912a-d5584c445e40',
    '362a3db9-96a2-5fb4-925b-c8500cf395c1',
    '85ca78e5-1b8f-5de2-a0af-41d06874d1d1',
    'Red Dead Redemption 2 для PS4',
    'Физическое издание на русском языке, состояние отличное.',
    'https://images.unsplash.com/photo-1542751371-adc38448a05e?auto=format&fit=crop&w=900&q=80',
    2490,
    'Псков',
    'active',
    '2026-08-04T16:05:00Z',
    '2026-08-04T16:05:00Z'
),
(
    'd4f45a72-f924-5fd5-98a1-6ab1ebcab104',
    'd3b90730-bf1f-5c12-95c7-b1ff3908167c',
    '85ca78e5-1b8f-5de2-a0af-41d06874d1d1',
    'Forza Horizon 5 для Xbox',
    'Лицензионный диск, полностью рабочий. Обмен на игры для PS5.',
    'https://images.unsplash.com/photo-1605901309584-818e25960a8f?auto=format&fit=crop&w=900&q=80',
    2990,
    'Псков',
    'active',
    '2026-08-03T11:20:00Z',
    '2026-08-03T11:20:00Z'
),
(
    '5792fb16-10c7-5701-9462-7df5b3cc983d',
    'f672cf38-b039-595d-84b3-821e9d5b3af7',
    '4a82b1a2-c73c-505e-91ba-e1cbc3b4ad25',
    'GeForce RTX 3070 Gaming OC 8 ГБ',
    'Рабочая видеокарта для игр в 2K. Проверка и самовывоз в Пскове.',
    'https://images.unsplash.com/photo-1591488320449-011701bb6704?auto=format&fit=crop&w=900&q=80',
    32990,
    'Псков',
    'active',
    '2026-08-02T14:10:00Z',
    '2026-08-02T14:10:00Z'
),
(
    'b337b8f3-49cf-5e4d-ba3a-4ad424cf256f',
    '5e96d7bb-c76c-5558-881e-1b132e49d342',
    '4a82b1a2-c73c-505e-91ba-e1cbc3b4ad25',
    'Видеокарта GTX 1660 Super',
    'Аккуратное состояние, работает стабильно. Возможен обмен.',
    '',
    16900,
    'Псков',
    'active',
    '2026-08-01T10:45:00Z',
    '2026-08-01T10:45:00Z'
),
(
    '2e32c450-be6d-585a-9223-479ac5b6b8d7',
    'f69088b0-d6a9-5f50-bb4e-be9b46cb8664',
    '85ca78e5-1b8f-5de2-a0af-41d06874d1d1',
    'God of War Ragnarök для PS5',
    'Физический диск, коробка без сколов. Можно обменять на другую игру.',
    'https://images.unsplash.com/photo-1605899435973-ca2d1a8861cf?auto=format&fit=crop&w=900&q=80',
    4490,
    'Псков',
    'active',
    '2026-07-31T19:20:00Z',
    '2026-07-31T19:20:00Z'
),
(
    'cfc4896d-6ef9-594c-91ca-cf9a0248886b',
    '1a9b30df-8e74-53f8-a55d-0c8a016995be',
    '85ca78e5-1b8f-5de2-a0af-41d06874d1d1',
    'GTA V для PS4',
    'Диск полностью рабочий, есть оригинальная коробка.',
    '',
    1990,
    'Псков',
    'active',
    '2026-07-30T15:00:00Z',
    '2026-07-30T15:00:00Z'
),
(
    '2cefe1a1-d372-5e2b-bf84-ac4dd877c793',
    'b6ead879-5ff3-5f09-af92-0003c53e4465',
    '85ca78e5-1b8f-5de2-a0af-41d06874d1d1',
    'Hogwarts Legacy для PS5',
    'Физическое издание в хорошем состоянии, один владелец.',
    'https://images.unsplash.com/photo-1542751371-adc38448a05e?auto=format&fit=crop&w=900&q=80',
    3590,
    'Псков',
    'active',
    '2026-07-29T11:35:00Z',
    '2026-07-29T11:35:00Z'
),
(
    'cf28bf4e-36ae-5c96-8a61-113f6c9f2a3a',
    '2db05252-81a6-5e50-b52f-57a19da8baa7',
    '85ca78e5-1b8f-5de2-a0af-41d06874d1d1',
    'Halo Infinite для Xbox',
    'Диск в хорошем состоянии. Рассмотрю обмен на игры для Xbox.',
    '',
    2290,
    'Псков',
    'active',
    '2026-07-28T09:15:00Z',
    '2026-07-28T09:15:00Z'
)
ON CONFLICT DO NOTHING;


-- Вишлисты товаров.
-- Владелец определяется через products.customer_id.
INSERT INTO wishlists (
    wishlist_id,
    product_id,
    name,
    created_at,
    updated_at
) VALUES
(
    '1ccf4307-7b00-5938-b425-42e30aee6f83',
    '1bb647e1-a136-5c68-9ad8-f7c3b880816b',
    'Что хочу получить за видеокарту',
    '2026-08-06T18:50:00Z',
    '2026-08-06T18:50:00Z'
),
(
    'cb3309cf-2dfb-5b40-b177-0f3ef5ba1973',
    '27d2d5c6-d819-51a8-a730-629f37d05784',
    'Обменяю на консоль или игры',
    '2026-08-05T12:30:00Z',
    '2026-08-05T12:30:00Z'
),
(
    '9a73882f-ffc5-5f66-b661-32acc48ba2eb',
    '5792fb16-10c7-5701-9462-7df5b3cc983d',
    'Что хочу получить за 3070',
    '2026-08-02T14:30:00Z',
    '2026-08-02T14:30:00Z'
),
(
    '38a9c2d4-ea36-5598-8a41-bc49cd4f708f',
    'd4f45a72-f924-5fd5-98a1-6ab1ebcab104',
    'Меняю на игры для PS5',
    '2026-08-03T11:40:00Z',
    '2026-08-03T11:40:00Z'
),
(
    'cec61017-f1e8-5212-90e2-7469615200cb',
    'cf28bf4e-36ae-5c96-8a61-113f6c9f2a3a',
    'Рассмотрю обмен на игры',
    '2026-07-28T09:40:00Z',
    '2026-07-28T09:40:00Z'
),
(
    'f6e0a11d-6db1-5b73-a008-cafaaa588b73',
    'a2adc3ce-b425-5671-912a-d5584c445e40',
    'Хочу видеокарту взамен',
    '2026-08-04T16:25:00Z',
    '2026-08-04T16:25:00Z'
),
(
    'b08d3402-803d-5715-addc-8668e65de1b3',
    '2cefe1a1-d372-5e2b-bf84-ac4dd877c793',
    'Меняю на комплектующие',
    '2026-07-29T11:50:00Z',
    '2026-07-29T11:50:00Z'
)
ON CONFLICT DO NOTHING;


-- Личные предпочтения пользователей.
-- Эти категории НЕ привязаны к конкретному товару или wishlist.
INSERT INTO customer_wishlist_options (customer_id, category_id) VALUES
('4f7a8183-d03c-52ad-9ef9-9821a1f40c8b', '85ca78e5-1b8f-5de2-a0af-41d06874d1d1'),

('549fe311-ecdd-5f4e-9c1d-cea2d100e286', '85ca78e5-1b8f-5de2-a0af-41d06874d1d1'),

('42bcc017-c6ef-5eb9-898f-6ff0d01293b2', '4a82b1a2-c73c-505e-91ba-e1cbc3b4ad25'),

('362a3db9-96a2-5fb4-925b-c8500cf395c1', '4a82b1a2-c73c-505e-91ba-e1cbc3b4ad25'),

('5e96d7bb-c76c-5558-881e-1b132e49d342', '85ca78e5-1b8f-5de2-a0af-41d06874d1d1'),

('f672cf38-b039-595d-84b3-821e9d5b3af7', '4a82b1a2-c73c-505e-91ba-e1cbc3b4ad25'),

('d3b90730-bf1f-5c12-95c7-b1ff3908167c', '85ca78e5-1b8f-5de2-a0af-41d06874d1d1'),

('f69088b0-d6a9-5f50-bb4e-be9b46cb8664', '85ca78e5-1b8f-5de2-a0af-41d06874d1d1'),

('1a9b30df-8e74-53f8-a55d-0c8a016995be', '85ca78e5-1b8f-5de2-a0af-41d06874d1d1'),

('b6ead879-5ff3-5f09-af92-0003c53e4465', '4a82b1a2-c73c-505e-91ba-e1cbc3b4ad25'),

('2db05252-81a6-5e50-b52f-57a19da8baa7', '85ca78e5-1b8f-5de2-a0af-41d06874d1d1')
ON CONFLICT DO NOTHING;


-- Категории, которые пользователь хочет получить за конкретный товар.
INSERT INTO wishlist_options (wishlist_id, category_id) VALUES
('1ccf4307-7b00-5938-b425-42e30aee6f83', '85ca78e5-1b8f-5de2-a0af-41d06874d1d1'),
('cb3309cf-2dfb-5b40-b177-0f3ef5ba1973', '85ca78e5-1b8f-5de2-a0af-41d06874d1d1'),
('9a73882f-ffc5-5f66-b661-32acc48ba2eb', '85ca78e5-1b8f-5de2-a0af-41d06874d1d1'),
('38a9c2d4-ea36-5598-8a41-bc49cd4f708f', '85ca78e5-1b8f-5de2-a0af-41d06874d1d1'),
('cec61017-f1e8-5212-90e2-7469615200cb', '85ca78e5-1b8f-5de2-a0af-41d06874d1d1'),
('f6e0a11d-6db1-5b73-a008-cafaaa588b73', '4a82b1a2-c73c-505e-91ba-e1cbc3b4ad25'),
('b08d3402-803d-5715-addc-8668e65de1b3', '4a82b1a2-c73c-505e-91ba-e1cbc3b4ad25')
ON CONFLICT DO NOTHING;


INSERT INTO chains (
    chain_id,
    from_product_id,
    to_product_id,
    initiator_id,
    recipient_id,
    status,
    message,
    expires_at,
    surcharge_amount,
    surcharge_currency,
    surcharge_payer,
    created_at,
    updated_at
) VALUES
(
    '8294f156-588c-5105-9113-2748be0be71a',
    'd4f45a72-f924-5fd5-98a1-6ab1ebcab104',
    'b337b8f3-49cf-5e4d-ba3a-4ad424cf256f',
    '5e96d7bb-c76c-5558-881e-1b132e49d342',
    'd3b90730-bf1f-5c12-95c7-b1ff3908167c',
    'completed',
    'Обменяю Forza Horizon 5 на GTX 1660 Super.',
    '2026-08-02T12:00:00Z',
    0,
    'RUB',
    NULL,
    '2026-08-01T12:00:00Z',
    '2026-08-02T11:30:00Z'
),
(
    'f66f7017-8307-5ebd-a9ed-3e8cd5ea1ddf',
    '1bb647e1-a136-5c68-9ad8-f7c3b880816b',
    '71aa4523-dca9-5f8e-9cb0-b448765d8c84',
    '4f7a8183-d03c-52ad-9ef9-9821a1f40c8b',
    '42bcc017-c6ef-5eb9-898f-6ff0d01293b2',
    'active',
    'Готов обменять RTX 3060 на Spider-Man 2 для PS5.',
    '2026-08-15T19:00:00Z',
    0,
    'RUB',
    NULL,
    '2026-08-06T19:00:00Z',
    '2026-08-06T19:00:00Z'
),
(
    '3485a4f0-8688-5017-a840-bedc7466f6c1',
    'a2adc3ce-b425-5671-912a-d5584c445e40',
    '27d2d5c6-d819-51a8-a730-629f37d05784',
    '362a3db9-96a2-5fb4-925b-c8500cf395c1',
    '549fe311-ecdd-5f4e-9c1d-cea2d100e286',
    'pending',
    'Рассмотрю обмен с доплатой.',
    '2026-08-14T13:00:00Z',
    0,
    'RUB',
    NULL,
    '2026-08-05T13:00:00Z',
    '2026-08-05T13:00:00Z'
)
ON CONFLICT DO NOTHING;


INSERT INTO chain_messages (
    message_id,
    chain_id,
    customer_id,
    body,
    created_at
) VALUES
(
    '74f4d113-4ed6-53fb-a447-809fd531578a',
    'f66f7017-8307-5ebd-a9ed-3e8cd5ea1ddf',
    '4f7a8183-d03c-52ad-9ef9-9821a1f40c8b',
    'Привет! Готов обсудить обмен RTX 3060 на Spider-Man 2.',
    '2026-08-06T19:05:00Z'
),
(
    '84a7b24a-ad89-5639-8d5d-fd1b216efaff',
    'f66f7017-8307-5ebd-a9ed-3e8cd5ea1ddf',
    '42bcc017-c6ef-5eb9-898f-6ff0d01293b2',
    'Да, давайте встретимся у ТЦ в субботу.',
    '2026-08-06T19:20:00Z'
)
ON CONFLICT DO NOTHING;


INSERT INTO chain_confirmations (
    chain_id,
    customer_id,
    success,
    reason,
    created_at
) VALUES
(
    '8294f156-588c-5105-9113-2748be0be71a',
    '5e96d7bb-c76c-5558-881e-1b132e49d342',
    TRUE,
    '',
    '2026-08-02T11:20:00Z'
),
(
    '8294f156-588c-5105-9113-2748be0be71a',
    'd3b90730-bf1f-5c12-95c7-b1ff3908167c',
    TRUE,
    '',
    '2026-08-02T11:30:00Z'
)
ON CONFLICT DO NOTHING;


INSERT INTO reviews (
    review_id,
    chain_id,
    from_customer_id,
    to_customer_id,
    product_id,
    rating,
    comment,
    created_at,
    updated_at
) VALUES
(
    'e8085ec7-a48e-59a5-bbc6-16ef1f1ae08b',
    '8294f156-588c-5105-9113-2748be0be71a',
    '5e96d7bb-c76c-5558-881e-1b132e49d342',
    'd3b90730-bf1f-5c12-95c7-b1ff3908167c',
    'b337b8f3-49cf-5e4d-ba3a-4ad424cf256f',
    5,
    'Быстро договорились, видеокарта соответствует описанию.',
    '2026-08-02T12:00:00Z',
    '2026-08-02T12:00:00Z'
),
(
    'b5e75ddb-10fb-5bef-b907-13f2e75a9d08',
    '8294f156-588c-5105-9113-2748be0be71a',
    'd3b90730-bf1f-5c12-95c7-b1ff3908167c',
    '5e96d7bb-c76c-5558-881e-1b132e49d342',
    'd4f45a72-f924-5fd5-98a1-6ab1ebcab104',
    4,
    'Всё хорошо, встретились в удобном месте и обменялись без проблем.',
    '2026-08-02T14:00:00Z',
    '2026-08-02T14:00:00Z'
)
ON CONFLICT DO NOTHING;

COMMIT;