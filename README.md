# Notion Hacks

A collection of tools to interact with notion from the command line.

## Installation

```
cd notionhacks
make
```

## Usage

1. Get your api key from notion
2. Store it securely in the keychain
  - `nn auth`
3. Give a name to a database
  - `nn register-db --name [my-name] --id [database-id]` (1)
4. Insert something
  - `nn insert --db tasks --name Foo -f Status=Inbox -f Area=PERSONAL`

--- 

_(1)_ For now, you need to manually run the following command to list your databases and get the id you want: 

```
curl 'https://api.notion.com/v1/databases' -H 'Authorization: Bearer [YOUR-API-KEY]' -H 'Notion-Version: 2021-05-13' | jq
```

