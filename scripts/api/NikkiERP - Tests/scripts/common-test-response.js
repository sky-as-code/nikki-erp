const tv4 = require('tv4');

module.exports.testHttpResponse = function (schema, httpStatus) {  
  test(`Status code is ${httpStatus}`, function () {
      expect(res.getStatus()).to.equal(httpStatus);
  });

  test("Response matches expected JSON schema", function () {
    // Uncomment for debugging. DO NOT COMMIT!
    // console.log({body: res.getBody()});
    const valid = tv4.validate(res.getBody(), schema);
    tv4.error && console.error(tv4.error);
    expect(tv4.error).not.to.exist;
    expect(valid).to.be.true;  
  });
};
