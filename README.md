# BitFunnel Lab Notebook

BitFunnel Lab Notebook is a simple _experiment framework_ for BitFunnel, an [experimental](http://bitfunnel.org/), [open source](https://github.com/BitFunnel/bitfunnel) fulltext search library. The goal of this project is to make it easy to create BitFunnel experiments that are _reproducible_, and in particular, which can be trivially run on as many platforms as possible.

Users declare an experiment by writing a YAML file that (among other things) specifies:

* The version of BitFunnel, and the specific configuration to be used to run an experiment.
* The version of the dataset was placed into BitFunnel for the experiment.
* The query log to be run against BitFunnel.

To run an experiment with precisely these settings, users need only pass this YAML to the LabBook binary:

```bash
go run src/main.go run $exptYaml $bitFunnelRoot $experimentRoot $corpusRoot
```

## Install

LabBook is written in pure Go, so getting the code is as simple as using `go get`:

```bash
go get -v github.com/BitFunnel/LabBook
```

Currently the code is in the `prototype` branch, so you still have to switch to the `LabBook` directory, and checkout the `prototype` branch before you can use `go run`.

## Build and run an experiment

You can see a hello world experiment [here](https://github.com/BitFunnel/Experiments/tree/master/hello-world).

Download your corpus, and put it in some folder, `$corpusRoot`. Decide which directory you'd like to put experiments in, and then go to the root of the repository, and run:

```bash
go run src/main.go run $exptYaml $bitFunnelRoot $experimentRoot $corpusRoot
```

This will complete all the steps required to run an experiment, from scratch, every time:

* Build BitFunnel.
  * Clone BitFunnel master to `$bitFunnelRoot` if it does not already exist, or if it does, we will fetch `origin master` in case the repository has been updated.
  * Check out the commit the experiment is meant to run on (specified in the YAML).
  * Run scripts to configure and build BitFunnel with `make`.
* Set up corpus.
  * Check the SHA of the tarballs in `$corpusRoot` if they exist; ensure they match with the SHAs that exist in the YAML. If they don't match or exist, error out.
  * Automatically create corpus manifest.
* Set up experiments.
  * Create `$experimentRoot` and `$experimentRoot/config` if they do not already exist. Else, continue.
  * Create samples of the corpus, which can be used to either configure BitFunnel, or ingested into BitFunnel for experiments.
  * Generate the term table and statistics from scratch, placing them in the config directory.
  * Fetch `queryLog` from a remote URL, verify the signature, and use it to Create an experiment script file. This script file will both run `verify` for each query (resulting in, among other things, the false positive rate), and then `query` for each one (more useful for things like timing experiments).
* Run experiment.
  * Launch the REPL
* Clean up.
  * Reset HEAD to the branch or revision it was at before we started.

## Declaring an experiment

Our YAML schema provides a declarative way of expressing an experiment, which the BFL can run on a wide variety of machines. There are 5 essential parts of each experiment:

1. The commit hash of BitFunnel that we expect the experiment to run on.
2. The corpus tarballs, containing data to ingest.
3. The query log, which contains queries to run in this experiment.
4. The corpus samples specifications, which describe how to sample the corpus. These samples are passed to the statistics package for configuration, and to the BitFunnel ingestion pipeline for the experiment itself.

In practice, this looks like this:

```yaml
bitfunnel-commit-hash: 618214be1ecd32afd09ccab1771e947d11cf6dab
lab-book-version: 0.0.1
query-log:
    raw-url: https://raw.githubusercontent.com/BitFunnel/Experiments/master/hello-world/queries.txt
    sha512: 56417da71c3373bcd49f8d22da9bc6af2ad4f0f4ee524572de884a85e571ab72c42aaea508b74c7cb98fc0aa948da854fd6747eead285ee4b75a84479f0b03db
corpus:
    - name: enwiki-20161020-chunked1.tar.gz
      sha512: 1a3be37650cbb6708c2c4385f6ebcf944d1cda933f2dd327de20acb6c72cf687737540f0108bcdcd4b6fc1e5014824bf1cdcb3304e87bfe6a82e0c7642b28e3f
samples:
    - name: configSample1
      gram-size: 1
      random-sample:
          seed: 42
          fraction: 0.2
      size-limits:
          min-posting-count: 50
          max-posting-count: 100
      max-documents: 5000000
    - name: ingestSample
      gram-size: 1
      random-sample:
          seed: 43
          fraction: 1
      size-limits:
          min-posting-count: 50
          max-posting-count: 100
statistics-config:
    gram-size: 1
    sample-name: configSample1
runtime-config:
    gram-size: 1
    ingest-threads: 1
    sample-name: ingestSample

```

## Proposed caching semantics

In the sections above, we describe a command that will run every step of the experiment from scratch, every time. This can be time-consuming, and it is useful to cache the costly steps that seldom change.

Below is a proposed workflow for managing cached data processing. Note in particular that we propose an additional CLI parameter, `$sampleRoot`; currently corpus samples are placed in the `$experimentRoot`, and the thought is that, with the ability to audit and trust the provinence of data, it should be possible to allow many experiments to use the same samples. Hence, we propose that samples are placed in some common root.

You may view the cache locking proposal [here](https://github.com/BitFunnel/LabBook/blob/prototype/src/experiment/file/lock/lock.go#L31-L48).

```bash
# Generates corpus samples. The fidelity of these samples are maintained by a
# LOCKFILE in each sample directory, containing a SHA512 of all the data files,
# which acts as the signature of the sample. Any configuration or BitFunnel
# experiment that depends on this version of the corpus samples will require
# this version of the corpus to run correctly.
go run src/main.go cache sample $exptYaml $bitFunnelRoot $corpusRoot $sampleRoot

# Generate configuration files (e.g., statistics, termtable, etc.). This
# configuration also contains a LOCKFILE, which contains both the SHA512 of the
# corpus it was generated from, as well as the 512 hash of the config data.
# Experiments that want to run on this cached data will check this hash to make
# sure they're correct.
go run src/main.go cache config $exptYaml $bitFunnelRoot $experimentRoot $sampleRoot

# Run the experiment. This directory contains a LOCKFILE that contains the
# SHA512 hash of the configuration and sample directories it depends on. If
# these don't match, we will error out.
go run src/main.go experiment run $exptYaml $bitFunnelRoot $experimentRoot $sampleRoot

# Locks a cache. Prevents us from overwriting it until we delete the lock file.
# We can lock an experiment corpus sample cache, a configuration cache, or an
# experiment cache.
go run src/main.go lock $expt_config_or_sample_directory
```

