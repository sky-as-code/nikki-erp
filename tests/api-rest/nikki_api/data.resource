*** Settings ***
Documentation     Unique test-data generators. Every value written to a unique column
...               must be built as {meaningful words} + {short hash} so that suites can
...               run repeatedly against the same database without collisions.
Library           String


*** Keywords ***
Unique Suffix
    ${suffix}=    Generate Random String    8    [LOWER][NUMBERS]
    RETURN    ${suffix}

Unique Email
    [Arguments]    ${words}=robot.user    ${domain}=e2e.nikki.vn
    ${suffix}=    Unique Suffix
    RETURN    ${words}.${suffix}@${domain}

Unique Display Name
    [Arguments]    ${words}=Robot Test User
    ${suffix}=    Unique Suffix
    RETURN    ${words} ${suffix}
