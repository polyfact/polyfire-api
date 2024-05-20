import argparse
import time
import sys
import os
import psycopg2
import hashlib
from urllib.parse import urlparse

MIGRATION_DIRECTORY=os.path.join(os.path.dirname(__file__), "../migrations")

parser = argparse.ArgumentParser(
                    prog='Migrate',
                    description='A simple database migration tool',
                    epilog='Developed for Polyfire')

subparsers = parser.add_subparsers()
parser_create = subparsers.add_parser('create', help="Create a new migration file in the migration directory")
parser_create.set_defaults(action="create")
parser_create.add_argument("name", help="The name of the new migration")

parser_migrate = subparsers.add_parser('migrate', help="Execute all the migrations from the migration directory that haven't been executed yet")
parser_migrate.set_defaults(action="migrate")
parser_migrate.add_argument("--ignore-rls", help="Indicate to the migrations they should manage the RLS security", action="store_false", dest="rls", default=True)

parser_rollback = subparsers.add_parser('rollback', help="Rollback the last migration batch")
parser_rollback.set_defaults(action="rollback")
parser_rollback.add_argument("-f", "--force", help="Force rollback to ignore diverging hash", action="store_true")
parser_rollback.add_argument("--ignore-rls", help="Indicate to the rollback they should manage the RLS security", action="store_false", dest="rls", default=True)

parser_init = subparsers.add_parser('init', help="Initialize the migration table")
parser_init.set_defaults(action="init")

result = parser.parse_args(sys.argv[1:])

postgres_urlparsed = urlparse(os.getenv("POSTGRES_URI"))
POSTGRES_USERNAME = postgres_urlparsed.username
POSTGRES_PASSWORD = postgres_urlparsed.password
POSTGRES_DATABASE = postgres_urlparsed.path[1:]
POSTGRES_HOSTNAME = postgres_urlparsed.hostname
POSTGRES_PORT = postgres_urlparsed.port
POSTGRES_CONN = psycopg2.connect(
    host=POSTGRES_HOSTNAME,
    database=POSTGRES_DATABASE,
    user=POSTGRES_USERNAME,
    password=POSTGRES_PASSWORD,
    port=POSTGRES_PORT
)

template = """def migrate(cur, rls=False):
    # cur.execute("SELECT 1")

def rollback(cur, rls=False):
    # cur.execute("SELECT 1")
"""

def create(name):
    file_name = "{}/{}-{}.py".format(MIGRATION_DIRECTORY, hex(int(time.time()))[2:], name)

    with open(file_name, 'w') as f:
        f.write(template)

    print("Created \"{}\"".format(file_name))

def insert_migration(batch, migration):
    with POSTGRES_CONN.cursor() as cur:
        cur.execute(
            "INSERT INTO migrations(id, file_name, batch, hash) VALUES (%s, %s, %s, %s)",
            (migration['id'],
            migration['file_name'],
            batch,
            migration['hash'],)
        )

def delete_migration(migration_id):
    with POSTGRES_CONN.cursor() as cur:
        cur.execute(
            "DELETE FROM migrations WHERE id = %s",
            (migration_id,)
        )

def execute_migration(file_name, rls=False):
    with POSTGRES_CONN.cursor() as cur:
        with open(file_name, 'r') as f:
            exec(f.read() + "\nmigrate(cur, rls=rls)", dict(
                cur=cur,
                __file__=file_name,
                rls=rls,
            ))

def execute_rollback(file_name, rls=False):
    with POSTGRES_CONN.cursor() as cur:
        with open(file_name, 'r') as f:
            exec(f.read() + "\nrollback(cur, rls=rls)", dict(
                cur=cur,
                __file__=file_name,
                rls=rls,
            ))

def file_hash(file_name):
    with open(file_name, 'rb') as f:
        digest = hashlib.file_digest(f, "sha256")
    return digest.hexdigest()

def list_migrations():
    migrations = [
        dict(
            id=x.split("-")[0],
            file_name=x,
            hash=file_hash("{}/{}".format(MIGRATION_DIRECTORY, x))
        )
        for x in os.listdir(MIGRATION_DIRECTORY)
        if x.endswith(".py")
    ]

    migrations.sort(key=lambda e: int(e['id'], 16))

    return migrations

def get_current_batch():
    with POSTGRES_CONN.cursor() as cur:
        cur.execute("SELECT COALESCE(MAX(batch), 0) FROM migrations")
        row = cur.fetchone()
        return row[0]

def get_previously_executed_migrations():
    with POSTGRES_CONN.cursor() as cur:
        cur.execute("SELECT id, file_name, batch, hash FROM migrations")
        return [dict(id=row[0], file_name=row[1], batch=row[2], hash=row[3]) for row in cur.fetchall()]

def get_new_migrations():
    all_migrations = { x['id']: x for x in list_migrations() }

    for m in get_previously_executed_migrations():
        if not m['id'] in all_migrations:
            raise ValueError('Missing migration "{}" (id: {})'.format(m['file_name'], m['id']))
        if m['hash'] != all_migrations[m['id']]['hash']:
            raise ValueError('Already executed migration "{}" (id: {}) and "{}" (id: {}) have the same id but their hash differs.\n("{}" != "{}")'.format(
                m['file_name'],
                m['id'],
                all_migrations[m['id']]['file_name'],
                all_migrations[m['id']]['id'],
                m['hash'],
                all_migrations[m['id']]['hash'],
            ))

        del all_migrations[m['id']]

    migrations = [x for x in all_migrations.values()]

    migrations.sort(key=lambda e: int(e['id'], 16))

    return migrations

def execute_new_migrations(rls=False):
    migrations = get_new_migrations()

    if len(migrations) == 0:
        print("No new migrations.")
        return

    batch = get_current_batch() + 1

    print("{} new migration{} found, executing batch {}...".format(len(migrations), "s" if len(migrations) > 1 else "", batch))
    for m in migrations:
        print("Executing new migration \"{}\"...".format(m['file_name']))
        execute_migration("{}/{}".format(MIGRATION_DIRECTORY, m['file_name']), rls=rls)
        insert_migration(batch, m)

    print("Done !")

def get_migration_batch(batch):
    with POSTGRES_CONN.cursor() as cur:
        cur.execute("SELECT id, file_name, batch, hash FROM migrations WHERE batch = %s", (batch,))
        return [dict(id=row[0], file_name=row[1], batch=row[2], hash=row[3]) for row in cur.fetchall()]

def get_batch_to_rollback(batch, force=False):
    all_migrations = { x['id']: x for x in list_migrations() }

    to_rollback = []

    for m in get_migration_batch(batch):
        if not m['id'] in all_migrations:
            raise ValueError('Missing migration "{}" (id: {})'.format(m['file_name'], m['id']))
        if m['hash'] != all_migrations[m['id']]['hash']:
            if force == False:
                raise ValueError('Migration "{}" (id: {}) and "{}" (id: {}) have the same id but their hash differs.\n("{}" != "{}")\nIf this is intentional, you can force the rollback with "--force".\nDO NOT FORCE THE ROLLBACK IF THIS MESSAGE IS UNEXPECTED ! THIS COULD LEAD TO A BAD STATE WHERE IT\'S IMPOSSIBLE TO KNOW WHICH MIGRATIONS HAVE BEEN EXECUTED OR NOT !'.format(
                    m['file_name'],
                    m['id'],
                    all_migrations[m['id']]['file_name'],
                    all_migrations[m['id']]['id'],
                    m['hash'],
                    all_migrations[m['id']]['hash'],
                ))
            else:
                print("Forcing rollback of migration with diverging hash (id: {})...", m['id'])

        to_rollback.append(all_migrations[m['id']])

    to_rollback.sort(key=lambda e: -int(e['id'], 16))

    return to_rollback

def rollback_batch(batch, force=False, rls=False):
    migrations = get_batch_to_rollback(batch, force)

    if len(migrations) == 0:
        print("Nothing to rollback.")
        return

    print("{} migration{} found in batch {}. Rollbacking batch {}...".format(len(migrations), "s" if len(migrations) > 1 else "", batch, batch))
    for m in migrations:
        print("Rollbacking migration \"{}\"...".format(m['file_name']))
        execute_rollback("{}/{}".format(MIGRATION_DIRECTORY, m['file_name']), rls=rls)
        delete_migration(m['id'])

    print("Done !")

def init():
    with POSTGRES_CONN.cursor() as cur:
        cur.execute("""
            CREATE TABLE migrations(
                id text primary key,
                file_name text,
                batch integer,
                hash text);

            CREATE EXTENSION IF NOT EXISTS "uuid-ossp" SCHEMA public;
            CREATE EXTENSION IF NOT EXISTS vector SCHEMA public;
        """)

try:
    match result.action:
        case "create":
            create(result.name)
        case "migrate":
            execute_new_migrations(rls=result.rls)
        case "rollback":
            rollback_batch(get_current_batch(), force=result.force, rls=result.rls)
        case "init":
            init()
        case _:
            raise Exeception("Unimplemented")

    POSTGRES_CONN.commit()
except Exception as error:
    print("ERROR:", error)
    if (result.action == "rollback"):
        print("An error occurred, all the queries executed during the rollback will be reverted...")
    else:
        print("An error occurred, rollback...")
    POSTGRES_CONN.rollback()
    POSTGRES_CONN.close()
    sys.exit(1)

POSTGRES_CONN.close()
sys.exit(0)
