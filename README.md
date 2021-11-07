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
3. Assign a name to a database
  - `nn register-db --name [my-name] --id [database-id]` (1)
4. Insert something
  - `nn insert --db tasks --name Foo -f Status=Inbox -f Area=PERSONAL`

--- 

_(1)_ For now, you need to manually run the following command to list your databases and get the id you want: 

## Available commands

- `nn list-db`: list all registered databases.
- `nn register-db --name $DB_NAME --id=$DB_ID`: assign a name to a database id to be used with other commands.
- `nn list --db=$DB_NAME`:  list all items in that database.
- `nn insert --db=$DB_NAME --name $TITLE -f $FIELD_1:$VALUE_1 -f $FIELD_2:$VALUE_2 ...`:  insert an item in the provided database.
- `nn open --db=$DB_NAME $PAGE_NAME`: open in a browser the first item found in that database that matches the provided title.
```
curl 'https://api.notion.com/v1/databases' -H 'Authorization: Bearer [YOUR-API-KEY]' -H 'Notion-Version: 2021-05-13' | jq
```

