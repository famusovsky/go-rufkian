# rufkian – backend

### Структура

Бэкенд состоит из 2х сервисов: companion и telephonist.

Код микросервисов построен в соответствии с трёхуровневой архитектурой:
- Корневой пакет содержит логику Транспортного уровня
- Подпакет database содержит логику уровня Базы данных
- Остальные подпакеты содержит логику Бизнес-уровня.

1. [**cmd**](./cmd/) – точки входа в приложения
2. [**internal**](./internal/) – внутренняя логика проекта
- [model](./internal/model/) – модели, используемые в обоих микросервисах
- [companion](./internal/companion/)
- - корень – реализация HTTP сервера
- - [database](./internal/companion/database/) – реализация клиента БД
- - [middleware](./internal/companion/middleware/) – middleware, отвечающие, например, за проверку пользователя
- - [proxy](./internal/companion/proxy/) – отвечает за отправку запросов в сторонние сервисы
- - [render](./internal/companion/render/) – вспомогательные функции для возвращения сгенерированного интерфейса
- - [auth](./internal/companion/auth/) – логика аутентификации пользователя
- - [dialog](./internal/companion/dialog/) – логика управления историей разговоров пользователя
- - [dictionary](./internal/companion/dictionary/) – содержит логика управления словарём пользователя
- - [user](./internal/companion/user/) – логика управления аккаунтом пользователя
- - [key](./internal/companion/key/) – логика отображения инструкции по добавлению ключа
- [telephonist](./internal/telephonist/)
- - корень – реализация HTTP сервера
- - [database](./internal/telephonist/database/) – реализация клиента БД
- - [walkietalkie](./internal/telephonist/walkietalkie/) – логика обработки текущих разговоров и получения ответов от LLM
- - [translator](./internal/telephonist/translator/) – логика взаимодействия с сервисом Яндекс.Переводчик
3. [**pkg**](./pkg/) – пакеты, использование которых возможно в сторонних проектах
- [***cookie***](./pkg/cookie/) – реализация безопасного управления файлами Cookie пользователя
- [***database***](./pkg/database/) – реализация подключения к sql БД
- [***grace***](./pkg/grace/) – реализация безопасной обработки экстренного завершения выполнения программы
- [***apkg***](./pkg/apkg/) – реализация генерации архива apkg для сервиса Anki
4. [**ui**](./ui/) – содержит файлы, описывающие интерфейс
- [views](./ui/views/) – шаблоны HTML
- [static](./ui/static/) – статичные файлы: иконки, css, javascript)

Файлы interface.go содержат определения интерфейсов, файлы const.go содержат
неизменяемые значения.
