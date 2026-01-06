import os
from pathlib import Path

from dotenv import load_dotenv
from pydantic_settings import BaseSettings

load_dotenv()


class EnvSettings(BaseSettings):
    """
    Global configuration and constants for mem.

    All paths and constants should be defined here to avoid scattered
    global variables throughout the codebase.
    """

    # Absolute path to the mem repo itself
    mem_working_dir: Path = Path(__file__).resolve().parent

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
    def mem_dir_stripped(self) -> str:
        """The .mem directory in the caller's project"""
        return ".mem/"

    @property
    def todos_dir(self) -> Path:
        """Directory for todo markdown files"""
        return self.mem_dir / "todos"

    @property
    def global_config_dir(self) -> Path:
        """Global mem config directory"""
        return Path.home() / ".config" / "mem"

    @property
    def global_config_file(self) -> Path:
        """Path to the global config.toml file"""
        return self.global_config_dir / "config.toml"

    @property
    def config_file(self) -> Path:
        """Path to the local config.toml file"""
        return self.mem_dir / "config.toml"

    @property
    def config_file_stripped(self) -> str:
        """Path to the config.toml file"""
        return ".mem/config.toml"

    @property
    def specs_dir(self) -> Path:
        """Directory for spec markdown files"""
        return self.mem_dir / "specs"

    @property
    def specs_dir_stripped(self) -> str:
        """Directory for spec markdown files"""
        return ".mem/specs"

    @property
    def logs_dir(self) -> Path:
        """Directory for work log files"""
        return self.mem_dir / "logs"

    @property
    def logs_dir_stripped(self) -> str:
        """Directory for work log files"""
        return ".mem/logs"


def get_env_settings():
    # make sure GITHUB_TOKEN is set
    assert os.getenv("GITHUB_TOKEN") is not None, "GITHUB_TOKEN must be set"
    return EnvSettings()


ENV_SETTINGS = get_env_settings()
