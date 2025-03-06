import alembic.util.exc
from alembic import command
from alembic.config import Config
from colorama import Style, Fore


def command_makemigrations():
    try:
        alembic_cfg = Config("alembic.ini")
        command.upgrade(alembic_cfg, "head")
        command.revision(alembic_cfg, message=None, autogenerate=True)
    except alembic.util.exc.CommandError:
        print('old migrations were found')
    else:
        print(Style.BRIGHT + Fore.GREEN + 'success')
