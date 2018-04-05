#!/usr/bin/env python3

import derision
import requests


def main():
    with derision.Client() as d:
        requests.post(
            url='http://localhost:5000/',
            data='a=x&a=y&b=z',
            headers={'Content-Type': 'application/x-www-form-urlencoded'},
        )

        assert d.get_requests()[0].form == {
            'a': ['x', 'y'],
            'b': ['z']
        }

    print('Done.')


if __name__ == '__main__':
    main()
