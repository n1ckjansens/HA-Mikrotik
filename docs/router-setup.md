# Настройка MikroTik RouterOS v7 для MikroTik Presence

Эта инструкция готовит роутер для `MikroTik Presence` Home Assistant add-on.

## Что важно заранее

- Используйте RouterOS **v7.12+** (рекомендуется последняя stable).
- REST API в RouterOS работает через сервисы `www`/`www-ssl`.
- Создайте **отдельного пользователя** только для Home Assistant.
- Ограничьте доступ по IP (только IP Home Assistant/Add-on).

## 1) Обновить RouterOS

```routeros
/system package update check-for-updates
/system package update install
```

После перезагрузки проверьте версию:

```routeros
/system resource print
```

## 2) Создать отдельного пользователя для HA

Пример для IP Home Assistant `192.168.88.2`:

```routeros
/user/group add name=ha-mikrotik-presence policy=read,test,web,api,rest-api
/user add name=ha_presence password="CHANGE_ME_STRONG_PASSWORD" group=ha-mikrotik-presence address=192.168.88.2/32 comment="Home Assistant MikroTik Presence"
```

Проверенный рабочий набор policy для REST в этом проекте:

```routeros
/user/group set [find name="ha-mikrotik-presence"] policy=read,test,web,api,rest-api
```

Проверка:

```routeros
/user print where name="ha_presence"
/user/group print where name="ha-mikrotik-presence"
```

## 3) Включить REST API (HTTPS рекомендуется)

### Вариант A (production): HTTPS (`www-ssl`)

Если уже есть валидный сертификат для роутера, используйте его.

Пример с локальным CA + серверным сертификатом:

```routeros
/certificate add name=ha-local-ca common-name=ha-local-ca key-usage=key-cert-sign,crl-sign
/certificate sign ha-local-ca
/certificate add name=router-web common-name=router.lan subject-alt-name=DNS:router.lan,IP:192.168.88.1 key-usage=digital-signature,key-encipherment,tls-server
/certificate sign router-web ca=ha-local-ca
```

Включить `www-ssl`, ограничить доступ только с HA IP:

```routeros
/ip service set www disabled=yes
/ip service set www-ssl disabled=no port=443 tls-version=only-1.2 certificate=router-web address=192.168.88.2/32
```

### Вариант B (только для локальной лаборатории): HTTP (`www`)

```routeros
/ip service set www disabled=no port=80 address=192.168.88.2/32
/ip service set www-ssl disabled=yes
```

## 4) Firewall: закрыть доступ извне

Ниже пример; адаптируйте под вашу текущую политику firewall.

```routeros
/ip firewall address-list add list=homeassistant address=192.168.88.2 comment="HA host"
/ip firewall filter add chain=input action=accept protocol=tcp dst-port=443 src-address-list=homeassistant comment="HA -> RouterOS REST"
/ip firewall filter add chain=input action=drop protocol=tcp dst-port=443 in-interface-list=WAN comment="Block REST from WAN"
```

Если используете HTTP в lab, замените `dst-port=443` на `dst-port=80`.

## 5) Проверить REST вручную

С машины Home Assistant (или любой разрешенной машины):

HTTPS (self-signed, без проверки сертификата):

```bash
curl -k -u ha_presence:CHANGE_ME_STRONG_PASSWORD https://192.168.88.1/rest/system/resource
```

Проверка обязательных endpoint-ов:

```bash
curl -k -u ha_presence:CHANGE_ME_STRONG_PASSWORD https://192.168.88.1/rest/ip/dhcp-server/lease
curl -k -u ha_presence:CHANGE_ME_STRONG_PASSWORD https://192.168.88.1/rest/interface/wifi/registration-table
curl -k -u ha_presence:CHANGE_ME_STRONG_PASSWORD https://192.168.88.1/rest/interface/bridge/host
curl -k -u ha_presence:CHANGE_ME_STRONG_PASSWORD https://192.168.88.1/rest/ip/arp
curl -k -u ha_presence:CHANGE_ME_STRONG_PASSWORD https://192.168.88.1/rest/ip/address
```

## 6) Заполнить конфигурацию add-on в Home Assistant

- `router_host`: `192.168.88.1` (или DNS имя/URL)
- `router_username`: `ha_presence`
- `router_password`: ваш пароль
- `router_ssl`: `true` для `www-ssl`, `false` для `www`
- `router_verify_tls`: `true`, если сертификат доверенный в HA; иначе `false`
- `poll_interval_sec`: 5-10 секунд (минимум 5)

## 7) Типовые проблемы

### 401/403 Unauthorized

- Неверный логин/пароль.
- Нет политики `rest-api` у пользователя.
- В `address` пользователя не включен IP Home Assistant.

### `{"detail":"std failure: not allowed (9)"}` на `/rest/system/resource`

- Логин успешный, но группе пользователя не хватает policy.
- Для этого проекта используйте:

```routeros
/user/group set [find name="ha-mikrotik-presence"] policy=read,test,web,api,rest-api
```

### timeout / cannot connect

- Неправильный `host`/порт.
- Сервис `www`/`www-ssl` выключен.
- Firewall блокирует вход на порт 80/443.

### TLS certificate verify failed

- Self-signed сертификат без доверия в HA.
- Временно поставьте `router_verify_tls=false` в конфигурации add-on.

### Wi-Fi таблица пустая

- На устройстве нет активных Wi-Fi клиентов.
- Убедитесь, что интерфейс Wi-Fi работает и клиенты подключены.

## Рекомендованный security baseline

- Только HTTPS (`www-ssl`), HTTP выключен.
- Отдельный пользователь `ha_presence` с `read,test,web,api,rest-api`.
- Доступ только с IP Home Assistant.
- Запрет доступа к REST с WAN.
- Регулярное обновление RouterOS до stable.

## Официальные источники MikroTik

- REST API: https://help.mikrotik.com/docs/spaces/ROS/pages/47579162/REST+API
- User/Policies: https://help.mikrotik.com/docs/spaces/ROS/pages/8978504/User
- Services (`www`, `www-ssl`): https://help.mikrotik.com/docs/spaces/ROS/pages/103841820/Services
- Certificates: https://help.mikrotik.com/docs/spaces/ROS/pages/2555969/Certificates
