# Тестовое задание ozon fintech на стажировку  

## Для запуска приложенние используйте:
```sh
# or
go run cmd/main.go

# or
docker-compose up
```

Переменные среды определяют тип хранилища и другие параметры

## Docker  

В Docker файле собирается приложение, после чего бинарник переносится в минимальный базовый образ.
## gqlgen - для работы с graphql

В итоге был выбран этот инструмент, ввиду того, что он сокращает кол-во бойлерплейт кода и избавляет от лишних ошибок по невнимательности

Так же эта библиотека не использует рефлексию и обеспечивает сильную типизацию.
## Проблемам N+1

Решением этой проблемы при работе с graphql - использование dataloader'а
## Пагинация взапросах

Пагинацию можно реализовать разными способами - я решил использовать курсор, руководствуясь рекомендациями [официальной докой graphql](https://graphql.org/learn/pagination/#pagination-and-edges)
