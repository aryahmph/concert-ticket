import {SharedArray} from "k6/data";
import http from "k6/http";
import {check, fail} from "k6";

const BaseURL = 'http://localhost:8080';

export const options = {
    summaryTrendStats: ['avg', 'p(75)', 'p(90)', 'p(95)'],
    scenarios: {
        concertTicket: {
            executor: 'constant-vus',
            vus: 2,
            duration: '30s',
        }
    },
    thresholds: {
        'http_req_duration{ListCategories:get}': [],
        'http_req_duration{CreateOrder:post}': [],
        'http_req_duration{PaymentCallback:post}': [],
    },
}

const data = new SharedArray('some name', function () {
    return JSON.parse(open('./emails.json')).emails;
});

export default function () {
    const categoriesRes = http.get(`${BaseURL}/categories`, {
        tags: {ListCategories: 'get'},
    });

    const checkListTickets = check(categoriesRes, {
        'is status OK': (r) => r.status === 200,
    });

    if (!checkListTickets) {
        console.log(`List tickets failed: ${categoriesRes.body}`)
        fail('Failed to list tickets');
    }

    const categories = categoriesRes.json('data');
    const randomCategories = Math.floor(Math.random() * categories.length);
    if (categories[randomCategories].total === 0) {
        console.log(`Category ${categories[randomCategories].name} is empty`)
        return;
    }

    // Create order
    const createOrderReq = {
        'email': data[Math.floor(Math.random() * data.length)],
        'category_id': categories[randomCategories].id,
    };

    const orderRes = http.post(`${BaseURL}/orders`,
        JSON.stringify(createOrderReq), {
            headers: {'Content-Type': 'application/json'},
            tags: {CreateOrder: 'post'},
        });

    check(orderRes, {
        'is status OK': (r) => r.status === 201 || r.status === 409,
    });

    if (orderRes.status === 500) {
        console.log(`Create order failed: ${orderRes.body}`)
        fail('Failed to create order');
    }

    if (orderRes.status >= 400) {
        return;
    }

    // Pay order
    if (__VU % 4 === 0) {
        return;
    }

    const order = orderRes.json('data');

    const payOrderReq = {
        'order_id': order.id,
        'status_code': '200',
        'transaction_status': 'settlement',
    };

    const payOrderRes = http.post(`${BaseURL}/payments/callback`,
        JSON.stringify(payOrderReq), {
            headers: {'Content-Type': 'application/json'},
            tags: {PaymentCallback: 'post'},
        });

    const checkPayOrder = check(payOrderRes, {
        'is status OK': (r) => r.status === 200,
    });

    if (!checkPayOrder) {
        console.log(`Pay order failed: ${payOrderRes.body}`)
        fail('Failed to pay order');
    }
}
