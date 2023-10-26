import os
import re

key_pattern = r'"from": "(0x[0-9a-fA-F]+)",'

def readfiles(dirname):
    new_keys = set()
    files = os.listdir(dirname)
    for file in files:
        if file.startswith('.'):
            continue
        print("Reading: ", os.path.join(dirname, file))
        with open(os.path.join(dirname, file), 'r') as f:
            for line in f:
                match = re.search(key_pattern, line)
                if match:
                    key = match.group(1)
                    new_keys.add(key)
    return new_keys

def main():
    keys = set()
    
    for batch_dir in os.listdir():
        if os.path.isdir(batch_dir):
            new_keys = readfiles(batch_dir)
            keys.update(new_keys)

    sorted_keys = sorted(list(keys))

    with open("keys.txt", 'w') as output_file:
        for key in sorted_keys:
            output_file.write(key + '\n')

    print("Extraction complete. Check keys.txt for the results.")

if __name__=='__main__':
    main()
