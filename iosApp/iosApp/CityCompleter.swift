import Foundation
import MapKit
import Observation

/// Wraps `MKLocalSearchCompleter` in an `@Observable` so SwiftUI can drive it
/// from a TextField without owning a delegate. Only city-shaped results are
/// surfaced (filter on `.address` plus a heuristic that the title looks like
/// a city, not a street).
///
/// Why not Apple's newer .pointOfInterestFilter API: that's iOS 16-only and
/// the result types are venues. Local search completer with the address
/// filter is the closest thing to a city autocomplete out of the box.
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
    }

    nonisolated func completerDidUpdateResults(_ completer: MKLocalSearchCompleter) {
        // Hop to MainActor — MapKit calls back on an arbitrary queue.
        let raw = completer.results
        Task { @MainActor in
            self.results = raw
                .map { CitySuggestion(title: $0.title, subtitle: $0.subtitle) }
                .filter { $0.looksLikeCity }
        }
    }

    nonisolated func completer(_ completer: MKLocalSearchCompleter, didFailWithError error: Error) {
        Task { @MainActor in self.results = [] }
    }
}

/// Lightweight wrapper so views don't import MapKit directly.
struct CitySuggestion: Identifiable, Hashable {
    let id = UUID()
    let title: String
    let subtitle: String

    /// MKLocalSearchCompleter returns streets and POIs too; we want city-ish
    /// rows. Heuristic: title has no number (street numbers, postal codes)
    /// and the subtitle looks like "Region, Country" or "Country".
    var looksLikeCity: Bool {
        let hasDigit = title.unicodeScalars.contains { CharacterSet.decimalDigits.contains($0) }
        if hasDigit { return false }
        // Reject results whose title contains a comma — those are usually
        // "Street, Neighborhood" rows. Cities are typically single-token.
        if title.contains(",") { return false }
        return !title.isEmpty
    }

    /// Best-effort country extraction. MKLocalSearchCompleter subtitles look
    /// like "State, Country" or just "Country". Take the trailing component.
    var country: String {
        subtitle.split(separator: ",").last.map { $0.trimmingCharacters(in: .whitespaces) } ?? ""
    }
}
