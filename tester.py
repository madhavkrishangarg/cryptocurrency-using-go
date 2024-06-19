import _sha256 as sha256

def test_sha256():
    print(sha256.sha256(b"hello").hexdigest() == "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824")



if __name__ == "__main__":
    test_sha256()