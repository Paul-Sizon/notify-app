import Foundation
import CoreLocation
import MapKit
import Observation

/// Wraps `MKLocalSearchCompleter` in an `@Observable` so SwiftUI can drive it
/// from a TextField without owning a delegate.
///
/// Two improvements over the naive completer:
/// 1. `region` is set to a worldwide span so famous cities rank above tiny
///    same-named villages near the device locale.
/// 2. Results are parsed into (city, region, country) and deduped by
///    (city + country), since Apple often returns both
///    "San Francisco / CA, United States" and "San Francisco, CA / United States".
@Observable
@MainActor
final class CityCompleter: NSObject, MKLocalSearchCompleterDelegate {
    var query: String = "" {
        didSet { completer.queryFragment = query }
    }
    private(set) var results: [CitySuggestion] = []

    private let completer = MKLocalSearchCompleter()

    override init() {
        super.init()
        completer.delegate = self
        completer.resultTypes = .address
        // Worldwide bias — without this MKLocalSearchCompleter ranks by
        // proximity to device region, which buries famous cities when the
        // device locale is far away (e.g. SF, CA missing for a CL device).
        completer.region = MKCoordinateRegion(
            center: CLLocationCoordinate2D(latitude: 0, longitude: 0),
            span: MKCoordinateSpan(latitudeDelta: 180, longitudeDelta: 360)
        )
    }

    nonisolated func completerDidUpdateResults(_ completer: MKLocalSearchCompleter) {
        let raw = completer.results.map { (title: $0.title, subtitle: $0.subtitle) }
        Task { @MainActor in
            var seen = Set<String>()
            var out: [CitySuggestion] = []
            for r in raw {
                guard let s = CitySuggestion.parse(title: r.title, subtitle: r.subtitle) else { continue }
                let key = "\(s.city.lowercased())|\(s.country.lowercased())"
                if seen.insert(key).inserted { out.append(s) }
            }
            self.results = out
        }
    }

    nonisolated func completer(_ completer: MKLocalSearchCompleter, didFailWithError error: Error) {
        Task { @MainActor in self.results = [] }
    }
}

/// Lightweight value type so views don't import MapKit / CoreLocation directly.
struct CitySuggestion: Identifiable, Hashable {
    let id = UUID()
    let city: String
    let region: String   // state/province, may be empty
    let country: String

    /// "City" line (top of row, also stored on the model).
    var title: String { city }

    /// "Region, Country" line. Drops empty region cleanly.
    var subtitle: String {
        region.isEmpty ? country : "\(region), \(country)"
    }

    /// Country-code → flag emoji. "" if unresolved.
    var flag: String { Self.flag(forCountry: country) }

    /// Parse Apple's `(title, subtitle)` pair into structured fields.
    /// Returns nil when the row clearly isn't a city (street, postal, etc).
    static func parse(title rawTitle: String, subtitle rawSub: String) -> CitySuggestion? {
        let title = rawTitle.trimmingCharacters(in: .whitespaces)
        let sub = rawSub.trimmingCharacters(in: .whitespaces)
        guard !title.isEmpty else { return nil }
        // Streets / postal codes have digits in the title.
        if title.unicodeScalars.contains(where: { CharacterSet.decimalDigits.contains($0) }) {
            return nil
        }

        // Title may be "San Francisco" or "San Francisco, CA".
        let titleParts = title.split(separator: ",").map { $0.trimmingCharacters(in: .whitespaces) }
        let city = titleParts.first ?? title
        let titleRegion = titleParts.count > 1
            ? titleParts.dropFirst().joined(separator: ", ")
            : ""

        // Subtitle may be "" (rare), "Country", or "Region, Country".
        // Treat the trailing component as the country and the rest as region.
        let subParts = sub.split(separator: ",").map { $0.trimmingCharacters(in: .whitespaces) }
        let country = subParts.last ?? ""
        let subRegion = subParts.count > 1
            ? subParts.dropLast().joined(separator: ", ")
            : ""

        // Reject if we have neither region nor country — usually means
        // the row is a country itself or a non-place result.
        if country.isEmpty && titleRegion.isEmpty { return nil }

        let region = !titleRegion.isEmpty ? titleRegion : subRegion
        return CitySuggestion(city: city, region: region, country: country)
    }

    private static func flag(forCountry name: String) -> String {
        // Resolve country name → ISO2 via the current locale's region table.
        // O(n) over ~250 regions — fine at autocomplete cadence.
        let n = name.lowercased()
        for code in Locale.Region.isoRegions.map(\.identifier) {
            if let localized = Locale.current.localizedString(forRegionCode: code)?.lowercased(),
               localized == n {
                return iso2ToFlag(code)
            }
        }
        return ""
    }

    private static func iso2ToFlag(_ code: String) -> String {
        guard code.count == 2 else { return "" }
        let base: UInt32 = 127397 // 0x1F1E6 - 'A'
        var s = ""
        for scalar in code.uppercased().unicodeScalars {
            if let u = UnicodeScalar(base + scalar.value) { s.unicodeScalars.append(u) }
        }
        return s
    }
}

// ─── current location ───

/// One-shot CoreLocation helper — request "when in use" auth, grab a single
/// fix, reverse-geocode it to a `CitySuggestion`. Designed for a single
/// `await finder.fetch()` call from the city screen.
@MainActor
final class CurrentLocationFinder: NSObject, CLLocationManagerDelegate {
    enum FinderError: Error { case denied, unavailable }

    private let manager = CLLocationManager()
    private var continuation: CheckedContinuation<CLLocation, Error>?

    override init() {
        super.init()
        manager.delegate = self
        manager.desiredAccuracy = kCLLocationAccuracyKilometer
    }

    func fetch() async throws -> CitySuggestion {
        let status = manager.authorizationStatus
        switch status {
        case .denied, .restricted: throw FinderError.denied
        case .notDetermined:
            manager.requestWhenInUseAuthorization()
            // Wait one runloop tick for the prompt to settle, then request.
            try await Task.sleep(nanoseconds: 100_000_000)
        default: break
        }

        let loc: CLLocation = try await withCheckedThrowingContinuation { cont in
            self.continuation = cont
            manager.requestLocation()
        }
        let placemarks = try await CLGeocoder().reverseGeocodeLocation(loc)
        guard let p = placemarks.first else { throw FinderError.unavailable }
        let city = p.locality ?? p.subAdministrativeArea ?? p.name ?? ""
        let region = p.administrativeArea ?? ""
        let country = p.country ?? ""
        guard !city.isEmpty else { throw FinderError.unavailable }
        return CitySuggestion(city: city, region: region, country: country)
    }

    nonisolated func locationManager(_ manager: CLLocationManager, didUpdateLocations locations: [CLLocation]) {
        Task { @MainActor in
            guard let loc = locations.last, let cont = self.continuation else { return }
            self.continuation = nil
            cont.resume(returning: loc)
        }
    }

    nonisolated func locationManager(_ manager: CLLocationManager, didFailWithError error: Error) {
        Task { @MainActor in
            guard let cont = self.continuation else { return }
            self.continuation = nil
            cont.resume(throwing: error)
        }
    }
}
