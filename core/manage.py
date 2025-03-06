from argparse import ArgumentParser
from importlib import import_module

from colorama import init


def main():
    parser = ArgumentParser(description="Database management commands")
    subparsers = parser.add_subparsers(dest="command")

    subparsers.add_parser('initialization', help='initialize databases')
    subparsers.add_parser('makemigrations', help='make migrations')
    subparsers.add_parser('migrate', help='apply migrations')

    args = parser.parse_args()
    if args.command == 'makemigrations':
        command = import_module('commands.makemigrations')
        command.command_makemigrations()
    elif args.command == 'migrate':
        command = import_module('commands.migrate')
        command.command_migrate()
    elif args.command == 'makemigrations':
        command = import_module('commands.initialization')
        command.command_initialization()

if __name__ == "__main__":
    init(autoreset=True)
    main()
