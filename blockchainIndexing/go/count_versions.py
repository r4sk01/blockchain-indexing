import os
import re

COUNT_FILE_1M = '1M-versions.txt'
COUNT_FILE_10M = '10M-versions.txt'
COUNT_FILE_50M = '50M-versions.txt'

KEY_PATTERN = r'"from": "(0x[0-9a-fA-F]+)",'

def read_files_in_dir(dirname):
    keys = {}
    files = os.listdir(dirname)
    for file in files:
        if file.startswith('.'):
            continue
        qualified_filename = os.path.join(dirname, file)
        new_keys = count_versions_in_file(qualified_filename)
        for key, value in new_keys.items():
                keys[key] = keys.get(key, 0) + value
    return keys

def count_versions_in_file(filename):
    print("Reading: ", filename)
    key_counts = {}
    with open(filename, 'r') as f:
        for line in f:
            match = re.search(KEY_PATTERN, line)
            if match:
                key = match.group(1)
                key_counts[key] = key_counts.get(key, 0) + 1
    return key_counts

def write_to_count_file(filename, key_counts):
    with open(filename, 'w') as output_file:
        for key, value in sorted(key_counts.items()):
            output_file.write(key + ' ' + str(value) + '\n')

def main():
    
    key_counts_1M = count_versions_in_file('./First100K/blockTransactions17000000-17010000.json')
    write_to_count_file(COUNT_FILE_1M, key_counts_1M)

    key_counts_10M = read_files_in_dir('./First100K')
    write_to_count_file(COUNT_FILE_10M, key_counts_10M)

    all_keys = key_counts_10M

    new_keys = read_files_in_dir('./Second100K')
    for key, value in new_keys.items():
        all_keys[key] = all_keys.get(key, 0) + value

    new_keys = read_files_in_dir('./Third100K')
    for key, value in new_keys.items():
        all_keys[key] = all_keys.get(key, 0) + value

    write_to_count_file(COUNT_FILE_50M, all_keys)

if __name__=='__main__':
    main()
