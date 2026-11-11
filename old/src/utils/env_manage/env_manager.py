import os
import secrets
import string
from typing import Optional

from src.utils.logger.logger import Log

TAG = "ENV_MANAGER"

class EnvManager:
    _instance = None
    _env_path = None
    _example_path = None

    def __new__(cls):
        if cls._instance is None:
            cls._instance = super(EnvManager, cls).__new__(cls)
            current_dir = os.path.dirname(os.path.abspath(__file__))
            project_root = os.path.dirname(os.path.dirname(os.path.dirname(current_dir)))
            cls._env_path = os.path.join(project_root, "config", ".env")
            cls._example_path = os.path.join(project_root, "config", ".env.example")
        return cls._instance

    @classmethod
    def init_env(cls):
        cls()
        
        if not os.path.exists(cls._env_path):
            Log.w(TAG, ".env file not found. Creating from .env.example...")
            if os.path.exists(cls._example_path):
                try:
                    with open(cls._example_path, 'r', encoding='utf-8') as example_file:
                        content = example_file.read()
                    with open(cls._env_path, 'w', encoding='utf-8') as env_file:
                        env_file.write(content)
                    Log.i(TAG, ".env file created successfully.")
                except Exception as e:
                    Log.e(TAG, "Failed to create .env file", error=e)
                    return
            else:
                Log.e(TAG, ".env.example file not found! Cannot create .env.")
                return
        cls._ensure_jwt_secret()

    @classmethod
    def _ensure_jwt_secret(cls):
        env_vars = cls._read_env_file()
        jwt_secret = env_vars.get("JWT_SECRET")

        if not jwt_secret or jwt_secret.strip() == "":
            Log.w(TAG, "JWT_SECRET is missing or empty. Generating a new secure secret...")
            new_secret = cls._generate_secure_secret()
            cls.set_env("JWT_SECRET", new_secret)
            Log.i(TAG, "JWT_SECRET generated and saved to .env.")

    @classmethod
    def _read_env_file(cls) -> dict:
        cls()
        
        env_vars = {}
        if not os.path.exists(cls._env_path):
            return env_vars

        try:
            with open(cls._env_path, 'r', encoding='utf-8') as f:
                for line in f:
                    line = line.strip()
                    if not line or line.startswith('#'):
                        continue
                    if '=' in line:
                        key, value = line.split('=', 1)
                        env_vars[key.strip()] = value.strip()
        except Exception as e:
            Log.e(TAG, "Failed to read .env file", error=e)
        return env_vars

    @classmethod
    def get_env(cls, key: str, default: Optional[str] = None) -> Optional[str]:
        env_vars = cls._read_env_file()
        return env_vars.get(key, default)

    @classmethod
    def set_env(cls, key: str, value: str):
        cls()
        
        if not os.path.exists(cls._env_path):
            Log.e(TAG, ".env file not found when trying to write.")
            return

        try:
            lines = []
            with open(cls._env_path, 'r', encoding='utf-8') as f:
                lines = f.readlines()

            key_found = False
            new_lines = []
            for line in lines:
                stripped_line = line.strip()
                if stripped_line.startswith('#') or not stripped_line:
                    new_lines.append(line)
                    continue
                
                if '=' in stripped_line:
                    current_key, _ = stripped_line.split('=', 1)
                    if current_key.strip() == key:
                        new_lines.append(f"{key}={value}\n")
                        key_found = True
                    else:
                        new_lines.append(line)
                else:
                    new_lines.append(line)

            if not key_found:
                if new_lines and not new_lines[-1].endswith('\n'):
                    new_lines.append('\n')
                new_lines.append(f"{key}={value}\n")

            with open(cls._env_path, 'w', encoding='utf-8') as f:
                f.writelines(new_lines)
            
        except Exception as e:
            Log.e(TAG, f"Failed to write {key} to .env file", error=e)

    @classmethod
    def _generate_secure_secret(cls, length=64) -> str:
        alphabet = string.ascii_letters + string.digits + "-_!@#$%^&*"
        return ''.join(secrets.choice(alphabet) for i in range(length))
