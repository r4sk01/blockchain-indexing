import os
import re

COUNT_FILES = ['1M-versions.txt', '10M-versions.txt', '50M-versions.txt']
MIN_ENTRIES = [1600, 2000, 3200]
ENTRY_LIMITS = [2000, 2500, 4000]
RANGE_SIZES = [10, 5, 5]

def read_count_file(filename):
    keys = {}
    with open(filename) as f:
        for line in f:
            key, versions = line.split()
            keys[key] = int(versions)
    return keys

def find_window(all_keys, target_sum, min_sum, range_size):
    
    sorted_keys = sorted(all_keys.items())

    current_sum = 0
    start_index = 0

    for end_index, (key, value) in enumerate(sorted_keys):
        current_sum += value

        if current_sum <= target_sum and current_sum >= min_sum and end_index - start_index + 1 == range_size:
            # print(f'RETURNING: Start: {start_index}, End: {end_index}, Sum: {current_sum}')
            return start_index

        # print(f'Window size: {end_index - start_index + 1}, Sum: {current_sum}')
            
        if end_index + 1 >= range_size:
            start_key, start_value = sorted_keys[start_index]
            current_sum -= start_value
            start_index += 1
        
        # print(f'Start: {start_index}, End: {end_index}, Sum: {current_sum}\n')

    return -1

def main():
    for i in range(3):
        keys = read_count_file(COUNT_FILES[i])
        start_index = find_window(keys, ENTRY_LIMITS[i], MIN_ENTRIES[i], RANGE_SIZES[i])
        sorted_keys = sorted(list(keys))
        if start_index >= 0:
            print(f'Range of {RANGE_SIZES[i]} keys containing at least {MIN_ENTRIES[i]} but less than {ENTRY_LIMITS[i]} total entries:')
            print(f'begins with key: {sorted_keys[start_index]} at index {start_index}')
        else:
            print(f'No range of size {RANGE_SIZES[i]} found with fewer than {ENTRY_LIMITS[i]} total entries')

if __name__=='__main__':
    main()
