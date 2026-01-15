import hashlib
import os
import hmac

class CryptographyManager:
    @staticmethod
    def get_password_hash(password: str) -> str:

        salt = os.urandom(16).hex()
        iterations = 100000
        
        hash_bytes = hashlib.pbkdf2_hmac(
            'sha256',
            password.encode('utf-8'),
            salt.encode('utf-8'),
            iterations
        )
        hash_hex = hash_bytes.hex()
        return f"{salt}${hash_hex}"

    @staticmethod
    def verify_password(plain_password: str, password_hash: str) -> bool:
        try:
            salt, hash_hex = password_hash.split('$')
            iterations = 100000
            
            verify_hash_bytes = hashlib.pbkdf2_hmac(
                'sha256',
                plain_password.encode('utf-8'),
                salt.encode('utf-8'),
                iterations
            )

            return hmac.compare_digest(verify_hash_bytes.hex(), hash_hex)
        except Exception:
            return False

crypto_manager = CryptographyManager()
