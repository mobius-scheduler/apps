from sklearn.gaussian_process import GaussianProcessRegressor
from sklearn.gaussian_process.kernels import RBF, ConstantKernel as C
import json, sys

# read JSON-ized input from stdin
inp = json.load(sys.stdin)

# init gaussian process kernel
#kernel = C(0.0799**2, (1e-5, 1e5)) * RBF([1.85,2.21], (1e-2, 1e2))
kernel = C(1.0, (1e-5, 1e5)) * RBF([1.0, 1.0], (1e-2, 1e2))
gp = GaussianProcessRegressor(kernel=kernel, optimizer=None, normalize_y=True)

# fit gp to inp
x = [(a['location']['latitude'], a['location']['longitude']) for a in inp['measurements']]
y = [a['aqi'] for a in inp['measurements']]
q = [(a['latitude'], a['longitude']) for a in inp['queries']]

gp.fit(x, y)
pred, sigma = gp.predict(q, return_std=True)

uncertain = []
for i in range(len(inp['queries'])):
    if q[i] in x:
        continue
    uncertainty = sigma[i]/pred[i]
    if uncertainty < inp['threshold']:
        uncertain += [inp['queries'][i]]

print(json.dumps(uncertain))

