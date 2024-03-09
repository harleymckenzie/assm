"""Setup configuration for the AWS Simple Session Manager package."""

from setuptools import setup, find_packages

setup(
    name='AWS Simple Session Manager',
    version='0.0.1',
    packages=find_packages(),
    entry_points={
        'console_scripts': [
            'assm=ssm-session:main',
        ],
    },
    install_requires=[
        'boto3',
        'simple-term-menu',
    ],
    python_requires='>=3.6',
    description=('A helper script to connect to AWS instances using '
                 'SSM Session Manager'),
    author='Harley McKenzie',
    author_email='mckenzie.harley@gmail.com',
    url='https://github.com/harleymckenzie/assm',
    classifiers=[
        'Programming Language :: Python :: 3',
        'License :: OSI Approved :: MIT License',
        'Operating System :: OS Independent',
    ],
)
