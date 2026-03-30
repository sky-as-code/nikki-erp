const { testHttpResponse } = require('./common-test-response');

module.exports.testSearch = function (size, page=0, itemCount=null) {  
  const schema = bru.getVar("search_schema");
  testHttpResponse(schema, 200);

  test(`Response is at page=${page} of size=${size} page`, function () {
      const payload = res.getBody();
      if (itemCount != null) {  
          expect(payload.items.length).to.equal(itemCount);
      } else {
          expect(payload.items.length).to.be.greaterThan(0);
      }
      expect(payload.size).to.equal(size);
      expect(payload.page).to.equal(page);
  });
};
