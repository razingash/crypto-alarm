from argparse import ArgumentParser
from importlib import import_module

from colorama import init


def main():
    parser = ArgumentParser(description="Database management commands")
    subparsers = parser.add_subparsers(dest="command")

    subparsers.add_parser('makemigrations', help='make migrations')
    subparsers.add_parser('initialization', help='initialize databases')
    subparsers.add_parser('deinitialization', help='deinitialize databases')

    args = parser.parse_args()
    if args.command == 'makemigrations':
        command = import_module('commands.makemigrations')
        command.command_makemigrations()
    elif args.command == 'initialization':
        command = import_module('commands.initialization')
        command.command_initialization()
    elif args.command == 'deinitialization':
        command = import_module('commands.deinitialization')
        command.command_deinitialization()


if __name__ == "__main__":
    init(autoreset=True)
    main()
