# Entropy Testing

In order verify that the random package actually generates random enough entropy/data, this test program holds the core functions that generate entropy as well as some noise makers to simulate a running program.

Please also note that output from `tickFeeder` is never used directly but fed as entropy to the actual RNG - `fortuna`.

With `tickFeeder`, to be sure that the delivered entropy is of high enough quality, only 1 bit of entropy is expected per generated byte - ie. we gather 8 times the amount we need. The following test below is run on the raw output.

To test the quality of entropy, first generate random data with the test program:

    go build

    ./test tickfeeder > output.bin # just the additional entropy feed
    # OR
    ./test fortuna > output.bin # the actual CSPRNG

    ls -lah output.bin # check filesize: should be ~1MB

Then, run `dieharder`, a random number generator test tool:

    dieharder -a -f output.bin

Below you can find two test outputs of `dieharder`.
Please note that around 5 tests of `dieharder` normally fail. This is expected and even desired.

`dieharder` output of 22KB of contextswitch (`go version go1.10.3 linux/amd64` on 23.08.2018):

    #=============================================================================#
    #            dieharder version 3.31.1 Copyright 2003 Robert G. Brown          #
    #=============================================================================#
       rng_name    |           filename             |rands/second|
            mt19937|                      output.bin|  1.00e+08  |
    #=============================================================================#
            test_name   |ntup| tsamples |psamples|  p-value |Assessment
    #=============================================================================#
       diehard_birthdays|   0|       100|     100|0.75124818|  PASSED  
          diehard_operm5|   0|   1000000|     100|0.71642114|  PASSED  
      diehard_rank_32x32|   0|     40000|     100|0.66406749|  PASSED  
        diehard_rank_6x8|   0|    100000|     100|0.79742497|  PASSED  
       diehard_bitstream|   0|   2097152|     100|0.68336079|  PASSED  
            diehard_opso|   0|   2097152|     100|0.99670345|   WEAK   
            diehard_oqso|   0|   2097152|     100|0.85930861|  PASSED  
             diehard_dna|   0|   2097152|     100|0.77857540|  PASSED  
    diehard_count_1s_str|   0|    256000|     100|0.27851730|  PASSED  
    diehard_count_1s_byt|   0|    256000|     100|0.29570009|  PASSED  
     diehard_parking_lot|   0|     12000|     100|0.51526020|  PASSED  
        diehard_2dsphere|   2|      8000|     100|0.49199324|  PASSED  
        diehard_3dsphere|   3|      4000|     100|0.99008122|  PASSED  
         diehard_squeeze|   0|    100000|     100|0.95518110|  PASSED  
            diehard_sums|   0|       100|     100|0.00015930|   WEAK   
            diehard_runs|   0|    100000|     100|0.50091086|  PASSED  
            diehard_runs|   0|    100000|     100|0.44091340|  PASSED  
           diehard_craps|   0|    200000|     100|0.77284264|  PASSED  
           diehard_craps|   0|    200000|     100|0.71027434|  PASSED  
     marsaglia_tsang_gcd|   0|  10000000|     100|0.38138922|  PASSED  
     marsaglia_tsang_gcd|   0|  10000000|     100|0.36661590|  PASSED  
             sts_monobit|   1|    100000|     100|0.06209802|  PASSED  
                sts_runs|   2|    100000|     100|0.82506539|  PASSED  
              sts_serial|   1|    100000|     100|0.99198615|  PASSED  
              sts_serial|   2|    100000|     100|0.85604831|  PASSED  
              sts_serial|   3|    100000|     100|0.06613657|  PASSED  
              sts_serial|   3|    100000|     100|0.16787860|  PASSED  
              sts_serial|   4|    100000|     100|0.45227401|  PASSED  
              sts_serial|   4|    100000|     100|0.43529092|  PASSED  
              sts_serial|   5|    100000|     100|0.99912474|   WEAK   
              sts_serial|   5|    100000|     100|0.94754128|  PASSED  
              sts_serial|   6|    100000|     100|0.98406523|  PASSED  
              sts_serial|   6|    100000|     100|0.92895983|  PASSED  
              sts_serial|   7|    100000|     100|0.45965410|  PASSED  
              sts_serial|   7|    100000|     100|0.64185152|  PASSED  
              sts_serial|   8|    100000|     100|0.57922926|  PASSED  
              sts_serial|   8|    100000|     100|0.52390292|  PASSED  
              sts_serial|   9|    100000|     100|0.82722325|  PASSED  
              sts_serial|   9|    100000|     100|0.89384819|  PASSED  
              sts_serial|  10|    100000|     100|0.79877889|  PASSED  
              sts_serial|  10|    100000|     100|0.49562348|  PASSED  
              sts_serial|  11|    100000|     100|0.09217966|  PASSED  
              sts_serial|  11|    100000|     100|0.00342361|   WEAK   
              sts_serial|  12|    100000|     100|0.60119444|  PASSED  
              sts_serial|  12|    100000|     100|0.20420318|  PASSED  
              sts_serial|  13|    100000|     100|0.76867489|  PASSED  
              sts_serial|  13|    100000|     100|0.35717970|  PASSED  
              sts_serial|  14|    100000|     100|0.67364089|  PASSED  
              sts_serial|  14|    100000|     100|0.98667204|  PASSED  
              sts_serial|  15|    100000|     100|0.24328833|  PASSED  
              sts_serial|  15|    100000|     100|0.52098866|  PASSED  
              sts_serial|  16|    100000|     100|0.48845863|  PASSED  
              sts_serial|  16|    100000|     100|0.61943558|  PASSED  
             rgb_bitdist|   1|    100000|     100|0.24694812|  PASSED  
             rgb_bitdist|   2|    100000|     100|0.75873723|  PASSED  
             rgb_bitdist|   3|    100000|     100|0.28670990|  PASSED  
             rgb_bitdist|   4|    100000|     100|0.41966273|  PASSED  
             rgb_bitdist|   5|    100000|     100|0.80463973|  PASSED  
             rgb_bitdist|   6|    100000|     100|0.44747725|  PASSED  
             rgb_bitdist|   7|    100000|     100|0.35848420|  PASSED  
             rgb_bitdist|   8|    100000|     100|0.56585089|  PASSED  
             rgb_bitdist|   9|    100000|     100|0.23179559|  PASSED  
             rgb_bitdist|  10|    100000|     100|0.83369283|  PASSED  
             rgb_bitdist|  11|    100000|     100|0.74761235|  PASSED  
             rgb_bitdist|  12|    100000|     100|0.50477673|  PASSED  
    rgb_minimum_distance|   2|     10000|    1000|0.29527530|  PASSED  
    rgb_minimum_distance|   3|     10000|    1000|0.83681186|  PASSED  
    rgb_minimum_distance|   4|     10000|    1000|0.85939646|  PASSED  
    rgb_minimum_distance|   5|     10000|    1000|0.90229335|  PASSED  
        rgb_permutations|   2|    100000|     100|0.99010460|  PASSED  
        rgb_permutations|   3|    100000|     100|0.99360922|  PASSED  
        rgb_permutations|   4|    100000|     100|0.30113906|  PASSED  
        rgb_permutations|   5|    100000|     100|0.60701235|  PASSED  
          rgb_lagged_sum|   0|   1000000|     100|0.37080580|  PASSED  
          rgb_lagged_sum|   1|   1000000|     100|0.91852932|  PASSED  
          rgb_lagged_sum|   2|   1000000|     100|0.74568323|  PASSED  
          rgb_lagged_sum|   3|   1000000|     100|0.64070201|  PASSED  
          rgb_lagged_sum|   4|   1000000|     100|0.53802729|  PASSED  
          rgb_lagged_sum|   5|   1000000|     100|0.67865656|  PASSED  
          rgb_lagged_sum|   6|   1000000|     100|0.85161494|  PASSED  
          rgb_lagged_sum|   7|   1000000|     100|0.37312323|  PASSED  
          rgb_lagged_sum|   8|   1000000|     100|0.17841759|  PASSED  
          rgb_lagged_sum|   9|   1000000|     100|0.85795513|  PASSED  
          rgb_lagged_sum|  10|   1000000|     100|0.79843176|  PASSED  
          rgb_lagged_sum|  11|   1000000|     100|0.21320830|  PASSED  
          rgb_lagged_sum|  12|   1000000|     100|0.94709672|  PASSED  
          rgb_lagged_sum|  13|   1000000|     100|0.12600611|  PASSED  
          rgb_lagged_sum|  14|   1000000|     100|0.26780352|  PASSED  
          rgb_lagged_sum|  15|   1000000|     100|0.07862730|  PASSED  
          rgb_lagged_sum|  16|   1000000|     100|0.21102254|  PASSED  
          rgb_lagged_sum|  17|   1000000|     100|0.82967141|  PASSED  
          rgb_lagged_sum|  18|   1000000|     100|0.05818566|  PASSED  
          rgb_lagged_sum|  19|   1000000|     100|0.01010140|  PASSED  
          rgb_lagged_sum|  20|   1000000|     100|0.17941782|  PASSED  
          rgb_lagged_sum|  21|   1000000|     100|0.98442639|  PASSED  
          rgb_lagged_sum|  22|   1000000|     100|0.30352772|  PASSED  
          rgb_lagged_sum|  23|   1000000|     100|0.56855155|  PASSED  
          rgb_lagged_sum|  24|   1000000|     100|0.27280405|  PASSED  
          rgb_lagged_sum|  25|   1000000|     100|0.41141889|  PASSED  
          rgb_lagged_sum|  26|   1000000|     100|0.25389013|  PASSED  
          rgb_lagged_sum|  27|   1000000|     100|0.10313177|  PASSED  
          rgb_lagged_sum|  28|   1000000|     100|0.76610028|  PASSED  
          rgb_lagged_sum|  29|   1000000|     100|0.97903830|  PASSED  
          rgb_lagged_sum|  30|   1000000|     100|0.51216732|  PASSED  
          rgb_lagged_sum|  31|   1000000|     100|0.98578832|  PASSED  
          rgb_lagged_sum|  32|   1000000|     100|0.95078719|  PASSED  
         rgb_kstest_test|   0|     10000|    1000|0.24930712|  PASSED  
         dab_bytedistrib|   0|  51200000|       1|0.51100031|  PASSED  
                 dab_dct| 256|     50000|       1|0.28794956|  PASSED  
    Preparing to run test 207.  ntuple = 0
            dab_filltree|  32|  15000000|       1|0.93283449|  PASSED  
            dab_filltree|  32|  15000000|       1|0.36488075|  PASSED  
    Preparing to run test 208.  ntuple = 0
           dab_filltree2|   0|   5000000|       1|0.94036105|  PASSED  
           dab_filltree2|   1|   5000000|       1|0.30118240|  PASSED  
    Preparing to run test 209.  ntuple = 0
            dab_monobit2|  12|  65000000|       1|0.00209003|   WEAK

`dieharder` of 1MB of fortuna (`go version go1.10.3 linux/amd64` on 23.08.2018):

    #=============================================================================#
    #            dieharder version 3.31.1 Copyright 2003 Robert G. Brown          #
    #=============================================================================#
       rng_name    |           filename             |rands/second|
            mt19937|                      output.bin|  8.44e+07  |
    #=============================================================================#
            test_name   |ntup| tsamples |psamples|  p-value |Assessment
    #=============================================================================#
       diehard_birthdays|   0|       100|     100|0.94302153|  PASSED  
          diehard_operm5|   0|   1000000|     100|0.08378380|  PASSED  
      diehard_rank_32x32|   0|     40000|     100|0.02062049|  PASSED  
        diehard_rank_6x8|   0|    100000|     100|0.43787871|  PASSED  
       diehard_bitstream|   0|   2097152|     100|0.15713023|  PASSED  
            diehard_opso|   0|   2097152|     100|0.79331996|  PASSED  
            diehard_oqso|   0|   2097152|     100|0.54138750|  PASSED  
             diehard_dna|   0|   2097152|     100|0.06957205|  PASSED  
    diehard_count_1s_str|   0|    256000|     100|0.21653644|  PASSED  
    diehard_count_1s_byt|   0|    256000|     100|0.96539542|  PASSED  
     diehard_parking_lot|   0|     12000|     100|0.21306362|  PASSED  
        diehard_2dsphere|   2|      8000|     100|0.40750466|  PASSED  
        diehard_3dsphere|   3|      4000|     100|0.99827314|   WEAK   
         diehard_squeeze|   0|    100000|     100|0.70994607|  PASSED  
            diehard_sums|   0|       100|     100|0.42729005|  PASSED  
            diehard_runs|   0|    100000|     100|0.08118125|  PASSED  
            diehard_runs|   0|    100000|     100|0.99226204|  PASSED  
           diehard_craps|   0|    200000|     100|0.49803401|  PASSED  
           diehard_craps|   0|    200000|     100|0.84011191|  PASSED  
     marsaglia_tsang_gcd|   0|  10000000|     100|0.40135552|  PASSED  
     marsaglia_tsang_gcd|   0|  10000000|     100|0.53311975|  PASSED  
             sts_monobit|   1|    100000|     100|0.96903259|  PASSED  
                sts_runs|   2|    100000|     100|0.55734041|  PASSED  
              sts_serial|   1|    100000|     100|0.69041819|  PASSED  
              sts_serial|   2|    100000|     100|0.61728694|  PASSED  
              sts_serial|   3|    100000|     100|0.70299864|  PASSED  
              sts_serial|   3|    100000|     100|0.36332027|  PASSED  
              sts_serial|   4|    100000|     100|0.57627216|  PASSED  
              sts_serial|   4|    100000|     100|0.95046929|  PASSED  
              sts_serial|   5|    100000|     100|0.79824554|  PASSED  
              sts_serial|   5|    100000|     100|0.62786166|  PASSED  
              sts_serial|   6|    100000|     100|0.84103529|  PASSED  
              sts_serial|   6|    100000|     100|0.89083859|  PASSED  
              sts_serial|   7|    100000|     100|0.69686380|  PASSED  
              sts_serial|   7|    100000|     100|0.79436099|  PASSED  
              sts_serial|   8|    100000|     100|0.84082295|  PASSED  
              sts_serial|   8|    100000|     100|0.95915719|  PASSED  
              sts_serial|   9|    100000|     100|0.48200567|  PASSED  
              sts_serial|   9|    100000|     100|0.10836112|  PASSED  
              sts_serial|  10|    100000|     100|0.45470523|  PASSED  
              sts_serial|  10|    100000|     100|0.97608829|  PASSED  
              sts_serial|  11|    100000|     100|0.89344380|  PASSED  
              sts_serial|  11|    100000|     100|0.31959825|  PASSED  
              sts_serial|  12|    100000|     100|0.43415812|  PASSED  
              sts_serial|  12|    100000|     100|0.27845148|  PASSED  
              sts_serial|  13|    100000|     100|0.50590833|  PASSED  
              sts_serial|  13|    100000|     100|0.39585514|  PASSED  
              sts_serial|  14|    100000|     100|0.55566778|  PASSED  
              sts_serial|  14|    100000|     100|0.57138798|  PASSED  
              sts_serial|  15|    100000|     100|0.12315118|  PASSED  
              sts_serial|  15|    100000|     100|0.41728831|  PASSED  
              sts_serial|  16|    100000|     100|0.23202389|  PASSED  
              sts_serial|  16|    100000|     100|0.84883373|  PASSED  
             rgb_bitdist|   1|    100000|     100|0.45137388|  PASSED  
             rgb_bitdist|   2|    100000|     100|0.93984739|  PASSED  
             rgb_bitdist|   3|    100000|     100|0.85148557|  PASSED  
             rgb_bitdist|   4|    100000|     100|0.77062397|  PASSED  
             rgb_bitdist|   5|    100000|     100|0.79511260|  PASSED  
             rgb_bitdist|   6|    100000|     100|0.86150140|  PASSED  
             rgb_bitdist|   7|    100000|     100|0.98572979|  PASSED  
             rgb_bitdist|   8|    100000|     100|0.73302973|  PASSED  
             rgb_bitdist|   9|    100000|     100|0.39660028|  PASSED  
             rgb_bitdist|  10|    100000|     100|0.13167592|  PASSED  
             rgb_bitdist|  11|    100000|     100|0.87937846|  PASSED  
             rgb_bitdist|  12|    100000|     100|0.80619403|  PASSED  
    rgb_minimum_distance|   2|     10000|    1000|0.38189429|  PASSED  
    rgb_minimum_distance|   3|     10000|    1000|0.21164619|  PASSED  
    rgb_minimum_distance|   4|     10000|    1000|0.91875064|  PASSED  
    rgb_minimum_distance|   5|     10000|    1000|0.27897081|  PASSED  
        rgb_permutations|   2|    100000|     100|0.22927506|  PASSED  
        rgb_permutations|   3|    100000|     100|0.80827585|  PASSED  
        rgb_permutations|   4|    100000|     100|0.38750474|  PASSED  
        rgb_permutations|   5|    100000|     100|0.18938169|  PASSED  
          rgb_lagged_sum|   0|   1000000|     100|0.72234187|  PASSED  
          rgb_lagged_sum|   1|   1000000|     100|0.28633796|  PASSED  
          rgb_lagged_sum|   2|   1000000|     100|0.52961866|  PASSED  
          rgb_lagged_sum|   3|   1000000|     100|0.99876080|   WEAK   
          rgb_lagged_sum|   4|   1000000|     100|0.39603203|  PASSED  
          rgb_lagged_sum|   5|   1000000|     100|0.01004618|  PASSED  
          rgb_lagged_sum|   6|   1000000|     100|0.89539065|  PASSED  
          rgb_lagged_sum|   7|   1000000|     100|0.55558774|  PASSED  
          rgb_lagged_sum|   8|   1000000|     100|0.40063365|  PASSED  
          rgb_lagged_sum|   9|   1000000|     100|0.30905028|  PASSED  
          rgb_lagged_sum|  10|   1000000|     100|0.31161899|  PASSED  
          rgb_lagged_sum|  11|   1000000|     100|0.76729775|  PASSED  
          rgb_lagged_sum|  12|   1000000|     100|0.36416009|  PASSED  
          rgb_lagged_sum|  13|   1000000|     100|0.21062168|  PASSED  
          rgb_lagged_sum|  14|   1000000|     100|0.17580591|  PASSED  
          rgb_lagged_sum|  15|   1000000|     100|0.54465457|  PASSED  
          rgb_lagged_sum|  16|   1000000|     100|0.39394806|  PASSED  
          rgb_lagged_sum|  17|   1000000|     100|0.81572681|  PASSED  
          rgb_lagged_sum|  18|   1000000|     100|0.98821505|  PASSED  
          rgb_lagged_sum|  19|   1000000|     100|0.86755786|  PASSED  
          rgb_lagged_sum|  20|   1000000|     100|0.37832948|  PASSED  
          rgb_lagged_sum|  21|   1000000|     100|0.52001140|  PASSED  
          rgb_lagged_sum|  22|   1000000|     100|0.83595676|  PASSED  
          rgb_lagged_sum|  23|   1000000|     100|0.22643336|  PASSED  
          rgb_lagged_sum|  24|   1000000|     100|0.96475696|  PASSED  
          rgb_lagged_sum|  25|   1000000|     100|0.49570837|  PASSED  
          rgb_lagged_sum|  26|   1000000|     100|0.71327165|  PASSED  
          rgb_lagged_sum|  27|   1000000|     100|0.07344404|  PASSED  
          rgb_lagged_sum|  28|   1000000|     100|0.86374872|  PASSED  
          rgb_lagged_sum|  29|   1000000|     100|0.24892548|  PASSED  
          rgb_lagged_sum|  30|   1000000|     100|0.14314375|  PASSED  
          rgb_lagged_sum|  31|   1000000|     100|0.27884009|  PASSED  
          rgb_lagged_sum|  32|   1000000|     100|0.66637341|  PASSED  
         rgb_kstest_test|   0|     10000|    1000|0.13954587|  PASSED  
         dab_bytedistrib|   0|  51200000|       1|0.54278716|  PASSED  
                 dab_dct| 256|     50000|       1|0.71177390|  PASSED  
    Preparing to run test 207.  ntuple = 0
            dab_filltree|  32|  15000000|       1|0.51006153|  PASSED  
            dab_filltree|  32|  15000000|       1|0.91162889|  PASSED  
    Preparing to run test 208.  ntuple = 0
           dab_filltree2|   0|   5000000|       1|0.15507188|  PASSED  
           dab_filltree2|   1|   5000000|       1|0.16787382|  PASSED  
    Preparing to run test 209.  ntuple = 0
            dab_monobit2|  12|  65000000|       1|0.28347219|  PASSED
