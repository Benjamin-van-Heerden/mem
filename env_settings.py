from pathlib import Path

from pydantic_settings import BaseSettings


class EnvSettings(BaseSettings):
    """
    Global configuration and constants for mem.

    All paths and constants should be defined here to avoid scattered
    global variables throughout the codebase.
    """

    # Absolute path to the mem repo itself
    mem_working_dir: Path = Path(__file__).resolve().parent.parent

    @property
    def migrations_dir(self) -> Path:
        """Absolute path to migrations in the mem repo itself"""
        return self.mem_working_dir / "migrations"

    @property
    def caller_dir(self) -> Path:
        """The directory where mem was invoked from (the project root)"""
        return Path.cwd()

    @property
    def mem_dir(self) -> Path:
        """The .mem directory in the caller's project"""
        return self.caller_dir / ".mem"

    @property
    def db_file(self) -> Path:
        """Path to the SQLite database"""
        return self.mem_dir / "mem.db"

    @property
    def config_file(self) -> Path:
        """Path to the config.toml file"""
        return self.mem_dir / "config.toml"

    @property
    def specs_dir(self) -> Path:
        """Directory for spec markdown files"""
        return self.mem_dir / "specs"

    @property
    def tasks_dir(self) -> Path:
        """Directory for task markdown files"""
        return self.mem_dir / "tasks"

    @property
    def logs_dir(self) -> Path:
        """Directory for work log files"""
        return self.mem_dir / "logs"


def get_env_settings():
    return EnvSettings()


ENV_SETTINGS = get_env_settings()
