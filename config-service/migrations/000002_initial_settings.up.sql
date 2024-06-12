insert into settings_services(service)
values ('api_gateway'),
       ('authentication_service'),
       ('config_service'),
       ('history_logs_consumer'),
       ('history_logs_service'),
       ('users_service'),
       ('registration_service'),
       ('email_service'),
       ('posts_service'),
       ('file_service');

insert into settings_items(service, key, value)
values ('api_gateway', 'ALLOWED_HOSTS', 'http://localhost:8000,http://localhost:4200,http://hidepost.ru'),
       ('api_gateway', 'AUTHENTICATION_SERVICE', 'authentication_service:8000'),
       ('api_gateway', 'CONFIG_SERVICE', 'config_service:8000'),
       ('api_gateway', 'HISTORY_LOGS_SERVICE', 'history_logs_service:8000'),
       ('api_gateway', 'USERS_SERVICE', 'users_service:8000'),
       ('api_gateway', 'HISTORY_LOGS_CONSUMER', 'history_logs_consumer:8000'),
       ('api_gateway', 'REGISTRATION_SERVICE', 'registration_service:8000'),
       ('api_gateway', 'EMAIL_SERVICE', 'email_service:8000'),
       ('api_gateway', 'POSTS_SERVICE', 'posts_service:8000');

insert into settings_items(service, key, value)
values ('authentication_service', 'JWT_SECRET', '123'),
       ('authentication_service', 'JWT_TOKEN_LIFETIME', '24'),
       ('authentication_service', 'DB_HOST', 'db'),
       ('authentication_service', 'DB_PORT', '5432'),
       ('authentication_service', 'DB_NAME', 'users'),
       ('authentication_service', 'DB_USER', 'postgres'),
       ('authentication_service', 'DB_PASSWORD', 'postgres'),
       ('authentication_service', 'LOG_QUEUE', 'logs'),
       ('authentication_service', 'MQ_HOST', 'mq'),
       ('authentication_service', 'MQ_PORT', '5672'),
       ('authentication_service', 'MQ_USER', 'rabbitmq'),
       ('authentication_service', 'MQ_PASSWORD', 'rabbitmq'),
       ('authentication_service', 'JWT_DEFAULT_LIFETIME', '1'),
       ('authentication_service', 'JWT_REMEMBER_ME_LIFETIME', '72');

insert into settings_items(service, key, value)
values ('config_service', 'DB_HOST', 'db'),
       ('config_service', 'DB_PORT', '5432'),
       ('config_service', 'DB_NAME', 'config'),
       ('config_service', 'DB_USER', 'postgres'),
       ('config_service', 'DB_PASSWORD', 'postgres'),
       ('config_service', 'LOG_QUEUE', 'logs'),
       ('config_service', 'MQ_HOST', 'mq'),
       ('config_service', 'MQ_PORT', '5672'),
       ('config_service', 'MQ_USER', 'rabbitmq'),
       ('config_service', 'MQ_PASSWORD', 'rabbitmq');

insert into settings_items(service, key, value)
values ('history_logs_consumer', 'DB_HOST', 'db'),
       ('history_logs_consumer', 'DB_PORT', '5432'),
       ('history_logs_consumer', 'DB_NAME', 'logs'),
       ('history_logs_consumer', 'DB_USER', 'postgres'),
       ('history_logs_consumer', 'DB_PASSWORD', 'postgres'),
       ('history_logs_consumer', 'LOG_QUEUE', 'logs'),
       ('history_logs_consumer', 'MQ_HOST', 'mq'),
       ('history_logs_consumer', 'MQ_PORT', '5672'),
       ('history_logs_consumer', 'MQ_USER', 'rabbitmq'),
       ('history_logs_consumer', 'MQ_PASSWORD', 'rabbitmq');

insert into settings_items(service, key, value)
values ('history_logs_service', 'DB_HOST', 'db'),
       ('history_logs_service', 'DB_PORT', '5432'),
       ('history_logs_service', 'DB_NAME', 'logs'),
       ('history_logs_service', 'DB_USER', 'postgres'),
       ('history_logs_service', 'DB_PASSWORD', 'postgres');

insert into settings_items(service, key, value)
values ('users_service', 'DB_HOST', 'db'),
       ('users_service', 'DB_PORT', '5432'),
       ('users_service', 'DB_NAME', 'users'),
       ('users_service', 'DB_USER', 'postgres'),
       ('users_service', 'DB_PASSWORD', 'postgres'),
       ('users_service', 'LOG_QUEUE', 'logs'),
       ('users_service', 'MQ_HOST', 'mq'),
       ('users_service', 'MQ_PORT', '5672'),
       ('users_service', 'MQ_USER', 'rabbitmq'),
       ('users_service', 'MQ_PASSWORD', 'rabbitmq'),
       ('users_service', 'WALLET_SERVICE_URL', 'http');

insert into settings_items(service, key, value)
values ('registration_service', 'DB_HOST', 'db'),
       ('registration_service', 'DB_PORT', '5432'),
       ('registration_service', 'DB_NAME', 'users'),
       ('registration_service', 'DB_USER', 'postgres'),
       ('registration_service', 'DB_PASSWORD', 'postgres'),
       ('registration_service', 'LOG_QUEUE', 'logs'),
       ('registration_service', 'MQ_HOST', 'mq'),
       ('registration_service', 'MQ_PORT', '5672'),
       ('registration_service', 'MQ_USER', 'rabbitmq'),
       ('registration_service', 'MQ_PASSWORD', 'rabbitmq'),
       ('registration_service', 'EMAIL_QUEUE', 'emails'),
       ('registration_service', 'USER_AUTO_ENABLE', 'false'),
       ('registration_service', 'SIGNUP_LIFETIME', '1'),
       ('registration_service', 'USER_CLEANUP_INTERVAL', '1'),
       ('registration_service', 'WALLET_SERVICE_URL', 'http');

insert into settings_items(service, key, value)
values ('email_service', 'LOG_QUEUE', 'logs'),
       ('email_service', 'MQ_HOST', 'mq'),
       ('email_service', 'MQ_PORT', '5672'),
       ('email_service', 'MQ_USER', 'rabbitmq'),
       ('email_service', 'MQ_PASSWORD', 'rabbitmq'),
       ('email_service', 'EMAIL_QUEUE', 'emails'),
       ('email_service', 'SMTP_BZ_API_KEY', ''),
       ('email_service', 'DEFAULT_EMAIL_ADDRESS', 'no-reply@hidepost.ru'),
       ('email_service', 'DEFAULT_EMAIL_DISPLAY_NAME', 'Хайд');

insert into settings_items(service, key, value)
values ('posts_service', 'DB_HOST', 'db'),
       ('posts_service', 'DB_PORT', '5432'),
       ('posts_service', 'DB_NAME', 'posts'),
       ('posts_service', 'DB_USER', 'postgres'),
       ('posts_service', 'DB_PASSWORD', 'postgres'),
       ('posts_service', 'LOG_QUEUE', 'logs'),
       ('posts_service', 'MQ_HOST', 'mq'),
       ('posts_service', 'MQ_PORT', '5672'),
       ('posts_service', 'MQ_USER', 'rabbitmq'),
       ('posts_service', 'MQ_PASSWORD', 'rabbitmq');

insert into settings_items(service, key, value)
values ('file_service', 'LOG_QUEUE', 'logs'),
       ('file_service', 'MQ_HOST', 'mq'),
       ('file_service', 'MQ_PORT', '5672'),
       ('file_service', 'MQ_USER', 'rabbitmq'),
       ('file_service', 'MQ_PASSWORD', 'rabbitmq'),
       ('file_service', 'FILE_QUEUE', 'files');