import argparse
import importlib

from colorama import init


def main():
    parser = argparse.ArgumentParser(description="Database management commands")
    subparsers = parser.add_subparsers(dest="command")

    subparsers.add_parser('makemigrations', help='make migrations')
    subparsers.add_parser('migrate', help='apply migrations')

    args = parser.parse_args()
    if args.command == 'makemigrations':
        command = importlib.import_module('commands.makemigrations')
        command.command_makemigrations()
    elif args.command == 'migrate':
        command = importlib.import_module('commands.migrate')
        command.command_migrate()

if __name__ == "__main__":
    init(autoreset=True)
    main()
