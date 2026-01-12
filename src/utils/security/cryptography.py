import hashlib
import os
import hmac

class CryptographyManager:
    @staticmethod
    def get_password_hash(password: str) -> str:
        """
        使用 PBKDF2_HMAC_SHA256 进行加盐哈希
        格式: salt$hash
        """
        salt = os.urandom(16).hex() # 生成 16 字节的随机盐
        iterations = 100000 # 迭代次数，越高越安全，但越慢
        
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
        """
        验证密码
        """
        try:
            salt, hash_hex = password_hash.split('$')
            iterations = 100000
            
            verify_hash_bytes = hashlib.pbkdf2_hmac(
                'sha256',
                plain_password.encode('utf-8'),
                salt.encode('utf-8'),
                iterations
            )
            
            # 使用 hmac.compare_digest 防止时序攻击
            return hmac.compare_digest(verify_hash_bytes.hex(), hash_hex)
        except Exception:
            return False

crypto_manager = CryptographyManager()
