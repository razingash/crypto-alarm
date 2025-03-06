from colorama import Style, Fore
from alembic import command
from alembic.config import Config

def command_migrate():
    try:
        alembic_cfg = Config("alembic.ini")
        command.upgrade(alembic_cfg, "head")
    except Exception as e:
        print(e)
    else:
        print(Style.BRIGHT + Fore.GREEN + 'migrations finished')
