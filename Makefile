## Добавить изменения и запушить (используйте: make commit MSG="ваше сообщение")
MIGRATE = docker compose run --rm migrate

#-------------------------------------------------------
#Миграции

#Добавление первой миграции
.PHONY: migrate-create
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Укажите NAME=имя_миграции"; \
		exit 1; \
	fi
	@cd backend/migrations && \
	last=$$(ls -1 [0-9]*_*.up.sql 2>/dev/null | tail -1 | cut -d'_' -f1 | sed 's/^0*//'); \
	if [ -z "$$last" ]; then \
		next=1; \
	else \
		next=$$((last + 1)); \
	fi; \
	padded=$$(printf "%06d" $$next); \
	touch "$${padded}_$(NAME).up.sql" "$${padded}_$(NAME).down.sql"; \
	echo "Созданы: $${padded}_$(NAME).up.sql и $${padded}_$(NAME).down.sql"

#Добавление новой версии миграции
.PHONY: migrate-up
migrate-up:
	$(MIGRATE) up

#Откат новой версии миграции
.PHONY: migrate-down
migrate-down:
	$(MIGRATE) down 1

#Принудительная установка версии
.PHONY: migrate-force
migrate-force:
	@if [ -z "$(VERSION)" ]; then \
		echo "Укажите VERSION=номер"; \
		exit 1; \
	fi
	$(MIGRATE) force $(VERSION)

#Текущая версия схемы миграций
.PHONY: migrate-status
migrate-status:
	$(MIGRATE) version
#------------------------------------------------------->
#Работа с докером

#Поднять контейнер
.PHONY: up
up:
	docker compose up -d --build

#Вырубить контейнеры
.PHONY: down
down:
	docker compose down -v

#Пересобрать докер
.PHONY: rebuild
rebuild:
	docker compose build
	docker compose up -d


#Логи контейнера
.PHONY: logs
logs:
	docker compose logs -f
#------------------------------------------------------->
#Линтеры и тесты

#Линтеры
.PHONY: lint
lint:
	docker run --rm -v $(PWD):/app -w /app/backend golangci/golangci-lint:latest golangci-lint run ./...

#Тесты
.PHONY: test
test:
	docker compose run --rm backend go test -v -race -cover ./...
#------------------------------------------------------->

#Быстрый коммит с сообщением
.PHONY: commit
commit:
	@if [ -z "$(MSG)" ]; then \
		echo "Ошибка: укажите сообщение коммита через MSG=\"текст\""; \
		exit 1; \
	fi
	git add -A
	git commit -m "$(MSG)"
	git push
	@echo "Изменения закоммичены и запушены!"

.PHONY: push
push: ## Просто запушить текущую ветку
	git push
	@echo "Пуш выполнен!"
